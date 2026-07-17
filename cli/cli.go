package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go-guestbook/models"
	"go-guestbook/msgops"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Run dispatches a CLI subcommand using the provided database connection.
func Run(db *gorm.DB, args []string) {
	if len(args) == 0 {
		PrintUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "users-seed":
		usersSeed(db)
	case "users-show":
		usersShow(db)
	case "users-clear":
		usersClear(db)
	case "messages-show":
		messagesShow(db)
	case "messages-seed":
		count := 10
		if len(args) > 1 {
			if _, err := fmt.Sscanf(args[1], "%d", &count); err != nil || count < 1 {
				log.Fatal().Str("count", args[1]).Msg("invalid message seed count")
			}
		}
		messagesSeed(db, count)
	case "messages-clear":
		messagesClear(db)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		PrintUsage()
		os.Exit(1)
	}
}

// PrintUsage writes supported CLI commands to standard output.
func PrintUsage() {
	fmt.Println(`Usage: guestbook <command>

Server:
  s, server               Start the HTTP server

Database:
  migrate                 Apply pending database migrations

User commands:
  users-seed              Create admin user (admin/admin); asks for confirmation
  users-show              List all users
  users-clear             Delete all users

Message commands:
  messages-show           List all messages
  messages-seed [count]   Generate sample messages (default: 10)
  messages-clear          Delete all messages`)
}

// usersSeed creates the default admin account when it does not already exist.
// It warns about the insecure admin/admin credentials and requires interactive confirmation.
// db is the open GORM connection used to look up and insert the user.
func usersSeed(db *gorm.DB) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("WARNING: Creating user with login \"admin\" and password \"admin\" is insecure and dangerous in production.")
	if !confirmYes(reader, "Do you want to continue? [y/N]: ") {
		fmt.Println("Aborted.")
		return
	}

	existing, err := models.FindUserByUsername(db, "admin")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check admin user")
	}
	if existing != nil {
		fmt.Println("User 'admin' already exists")
		return
	}

	user, err := models.CreateUser(db, "admin", "admin")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create admin user")
	}
	fmt.Printf("Created user: id=%d username=%s\n", user.ID, user.Username)
}

// confirmYes prompts on stdin and reports whether the answer is an affirmative yes.
// reader is the buffered stdin reader; prompt is the text shown before reading a line.
func confirmYes(reader *bufio.Reader, prompt string) bool {
	answer := strings.ToLower(readLine(reader, prompt))
	return answer == "y" || answer == "yes"
}

// usersShow prints all administrator accounts.
func usersShow(db *gorm.DB) {
	var users []models.User
	if err := db.Order("id asc").Find(&users).Error; err != nil {
		log.Fatal().Err(err).Msg("failed to fetch users")
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("%-5s %-20s %-20s\n", "ID", "Username", "Created At")
	fmt.Println(strings.Repeat("-", 47))
	for _, user := range users {
		fmt.Printf("%-5d %-20s %-20s\n", user.ID, user.Username, user.CreatedAt.Format("2006-01-02 15:04"))
	}
}

// usersClear deletes every administrator account.
func usersClear(db *gorm.DB) {
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.User{})
	if result.Error != nil {
		log.Fatal().Err(result.Error).Msg("failed to clear users")
	}
	fmt.Printf("Deleted %d user(s).\n", result.RowsAffected)
}

// messagesShow prints all guestbook messages.
func messagesShow(db *gorm.DB) {
	var messages []models.Message
	if err := db.Order("id asc").Find(&messages).Error; err != nil {
		log.Fatal().Err(err).Msg("failed to fetch messages")
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return
	}

	fmt.Printf("%-5s %-20s %-25s %s\n", "ID", "Author", "Created At", "Content")
	fmt.Println(strings.Repeat("-", 80))
	for _, message := range messages {
		content := message.Content
		if len(content) > 40 {
			content = content[:40] + ".."
		}
		fmt.Printf("%-5d %-20s %-25s %s\n",
			message.ID,
			message.Author,
			message.CreatedAt.Format("2006-01-02 15:04"),
			content,
		)
	}
}

// messagesSeed inserts sample guestbook messages.
func messagesSeed(db *gorm.DB, count int) {
	created, err := msgops.Seed(db, count)
	if err != nil {
		log.Fatal().Err(err).Int("created", created).Msg("failed to seed messages")
	}
	fmt.Printf("Created %d message(s).\n", created)
}

// messagesClear deletes every guestbook message.
func messagesClear(db *gorm.DB) {
	deleted, err := msgops.Clear(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to clear messages")
	}
	fmt.Printf("Deleted %d message(s).\n", deleted)
}

// readLine reads a single trimmed line from standard input.
func readLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read input")
	}
	return strings.TrimSpace(line)
}
