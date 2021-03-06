package main

import (
	"log"
	"time"

	"github.com/robfig/cron"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/mohuishou/scuplus-go/job"
	"github.com/mohuishou/scuplus-go/model"
)

const pageSize = 200

func main() {
	c := cron.New()
	// 每天 9,12,21点分别执行一次
	c.AddFunc("0 0 9,12,21 1/1 * ? ", func() {
		updateAll()
	})
	// 每天晚上八点执行一次
	c.AddFunc("0 0 20 1/1 * ? ", func() {
		//book()
	})
	// 每天早上8点执行一次
	c.AddFunc("0 0 8 1/1 * ? ", func() {
		//exam()
	})
	// 每天凌晨1点执行一次
	c.AddFunc("0 0 1 1/1 * ? ", func() {
		courseCountEvaluate()
		calCourse()
	})
	c.Start()
	select {}
}

func updateAll() {
	count := 0
	model.DB().Table("users").Where("jwc_verify = 1").Count(&count)
	log.Println("教务处绑定用户总计：", count)
	for i := 0; i < (count/pageSize + 1); i++ {
		log.Printf("第%d批，每批%d", i, pageSize)
		users := []model.User{}
		model.DB().Where("jwc_verify = 1").Select([]string{"id"}).Offset((i - 1) * pageSize).Limit(pageSize).Find(&users)
		for _, user := range users {
			sign := &tasks.Signature{
				Name: "update_all",
				Args: []tasks.Arg{
					{
						Type:  "uint",
						Value: user.ID,
					},
				},
			}

			_, err := job.Server.SendTask(sign)
			if err != nil {
				log.Println("cron error update all", err)
			}
		}
	}
}

func book() {
	count := 0
	model.DB().Table("users").Where("verify = 1").Count(&count)
	for i := 0; i < (count/pageSize + 1); i++ {
		users := []model.User{}
		model.DB().Where("verify = 1").Select([]string{"id"}).Offset((i - 1) * pageSize).Limit(pageSize).Find(&users)
		for _, user := range users {
			// 获取即将到期的书籍
			now := time.Now()
			var book model.LibraryBook
			model.DB().Where("user_id = ? and due_time > ?", user.ID, now).Find(&book)

			day := (book.DueTime.Unix() - now.Unix()) / 3600 / 24
			if day == 7 || day == 1 {
				sign := &tasks.Signature{
					Name: "notify_book",
					Args: []tasks.Arg{
						{
							Type:  "uint",
							Value: user.ID,
						},
						{
							Type:  "string",
							Value: book.Title,
						},
						{
							Type:  "string",
							Value: book.DueDate,
						},
						{
							Type:  "int64",
							Value: day,
						},
					},
				}

				_, err := job.Server.SendTask(sign)
				if err != nil {
					log.Println("cron error book", err)
				}
			}
		}
	}
}

func exam() {
	count := 0
	model.DB().Table("users").Where("verify = 1").Count(&count)
	for i := 0; i < (count/pageSize + 1); i++ {
		users := []model.User{}
		model.DB().Where("jwc_verify = 1").Select([]string{"id"}).Offset((i - 1) * pageSize).Limit(pageSize).Find(&users)
		for _, user := range users {
			// 获取最近的考试
			now := time.Now()
			var exam model.Exam
			model.DB().Where("user_id = ? and start_time > ?", user.ID, now).Find(&exam)
			day := (exam.StartTime.Unix() - now.Unix()) / 3600 / 24
			if day == 30 || day == 7 || day == 0 {
				sign := &tasks.Signature{
					Name: "notify_exam",
					Args: []tasks.Arg{
						{
							Type:  "uint",
							Value: user.ID,
						},
						{
							Type:  "string",
							Value: exam.CourseName,
						},
						{
							Type:  "string",
							Value: exam.Date,
						},
						{
							Type:  "string",
							Value: exam.Time,
						},
						{
							Type:  "string",
							Value: exam.Campus + exam.Building + exam.Classroom,
						},
						{
							Type:  "string",
							Value: exam.Site,
						},
						{
							Type:  "string",
							Value: exam.Name,
						},
						{
							Type:  "int64",
							Value: day,
						},
					},
				}
				_, err := job.Server.SendTask(sign)
				if err != nil {
					log.Println("cron error exam", err)
				}
			}
		}
	}
}
