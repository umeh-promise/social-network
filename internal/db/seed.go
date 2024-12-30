package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/umeh-promise/social/internal/store"
)

var usernames = []string{"user01", "coolCat", "techGuru", "skyWalker", "happyBee",
	"ninjaCoder", "swiftHawk", "pixelPanda", "cryptoKing", "oceanBliss",
	"codeMaster", "javaJunkie", "frontendFan", "backendBoss", "dataDiver",
	"rainyDay", "starGazer", "zenMode", "devDragon", "logicLover",
	"byteWizard", "apiExplorer", "designDiva", "spaceCoder", "fastFinger",
	"buildBro", "htmlHero", "cssCrusader", "bugHunter", "reactRanger",
	"nodeNinja", "mongoMaverick", "sqlSniper", "graphQLGuru", "scriptSage",
	"cloudChaser", "deployDynamo", "unitTester", "integrationAce", "uiWarrior",
	"darkCoder", "neoNerd", "matrixMaverick", "zeroCool", "cyberPunk",
	"wildWolf", "cleverKoala", "superPenguin", "phantomByte", "quantumQuest"}

var titles = []string{"The Future of AI: What Lies Ahead?",
	"10 Productivity Hacks for Developers",
	"Understanding the Basics of Blockchain Technology",
	"How to Build a Personal Brand as a Software Engineer",
	"Top 5 JavaScript Frameworks to Learn in 2024",
	"The Role of UX in Modern Web Design",
	"How to Secure Your APIs: Best Practices",
	"Remote Work: Tips for Staying Motivated and Connected",
	"Mastering Git: Advanced Techniques for Version Control",
	"The Rise of No-Code Tools: Will Developers Become Obsolete?",
	"Demystifying Cloud Computing for Beginners",
	"Top 10 Mistakes to Avoid When Learning to Code",
	"How to Land Your First Job in Tech Without a Degree",
	"The Art of Writing Clean and Maintainable Code",
	"Why Cybersecurity is Everyone’s Responsibility",
	"How to Design Accessible Websites: A Practical Guide",
	"AI in Healthcare: Opportunities and Challenges",
	"The Importance of Continuous Learning in Tech Careers",
	"Building Scalable Applications: Lessons Learned",
	"Exploring the Latest Trends in Mobile App Development"}

var tags = []string{"AI", "Future Trends", "Technology", "Machine Learning", "Innovation"}

var comments = []string{"AI is evolving so fast; it’s both exciting and scary!",
	"Great insights on the future of AI. Looking forward to more.",
	"I wonder how ethical considerations will shape AI development."}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating users:", err)
			return
		}
	}
	tx.Commit()

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating posts:", err)
			return
		}
	}

	comments := generateComments(300, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comments", err)
		}
	}
	log.Println("Seed completed")

}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Role: store.Role{
				Name: "user",
			},
		}
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	for i := 0; i < num; i++ {
		user := users[rand.IntN(len(users))]

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.IntN(len(titles))],
			Content: comments[rand.IntN(len(comments))],
			Tags: []string{
				tags[rand.IntN(len(tags))],
				tags[rand.IntN(len(tags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)

	for i := 0; i < num; i++ {
		cms[i] = &store.Comment{
			UserID:  users[rand.IntN(len(users))].ID,
			PostID:  posts[rand.IntN(len(posts))].ID,
			Content: comments[rand.IntN(len(comments))],
		}
	}
	return cms
}
