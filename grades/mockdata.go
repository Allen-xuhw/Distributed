package grades

func init()  {
	students = []Student{
		{
			ID: 1,
			FirstName: "Nick",
			LastName: "Carter",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type: GradeQuiz,
					Score: 85,
				},
				{
					Title: "Final Exam",
					Type: GradeExam,
					Score: 94,
				},
				{
					Title: "Quiz 2",
					Type: GradeQuiz,
					Score: 82,
				},
			},
		},
	}
}
