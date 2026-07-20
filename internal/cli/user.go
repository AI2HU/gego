package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/auth"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
)

var (
	userUsername string
	userPassword string
	userRole     string
	userID       string
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage API users",
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API user",
	RunE:  runUserCreate,
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API users",
	RunE:  runUserList,
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an API user",
	RunE:  runUserDelete,
}

func init() {
	userCreateCmd.Flags().StringVar(&userUsername, "username", "", "Username for the new user")
	userCreateCmd.Flags().StringVar(&userPassword, "password", "", "Password for the new user")
	userCreateCmd.Flags().StringVar(&userRole, "role", "member", "Role for the new user (admin or member)")

	userDeleteCmd.Flags().StringVar(&userID, "id", "", "User ID to delete")
	userDeleteCmd.Flags().StringVar(&userUsername, "username", "", "Username to delete")

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userDeleteCmd)
	rootCmd.AddCommand(userCmd)
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := runDatabaseMigrations(ctx, database); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if userUsername == "" {
		return fmt.Errorf("--username is required")
	}

	role := models.Role(userRole)
	if !role.Valid() {
		return fmt.Errorf("invalid role: %s (must be admin or member)", userRole)
	}

	password := userPassword
	if password == "" {
		reader := bufio.NewReader(os.Stdin)
		var err error
		password, err = promptWithRetry(reader, "Password (min 8 characters): ", func(input string) (string, error) {
			if len(input) < 8 {
				return "", fmt.Errorf("password must be at least 8 characters")
			}
			return input, nil
		})
		if err != nil {
			return err
		}
	}

	authService := services.NewAuthService(database, auth.Config{})
	user, err := authService.CreateUser(ctx, userUsername, password, role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("✅ Created user %s with role %s\n", user.Username, user.Role)
	return nil
}

func runUserList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := runDatabaseMigrations(ctx, database); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	authService := services.NewAuthService(database, auth.Config{})
	users, err := authService.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Printf("%-36s %-20s %-10s %s\n", "ID", "USERNAME", "ROLE", "CREATED AT")
	fmt.Println(strings.Repeat("-", 90))
	for _, user := range users {
		fmt.Printf("%-36s %-20s %-10s %s\n",
			user.ID,
			user.Username,
			user.Role,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	return nil
}

func runUserDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := runDatabaseMigrations(ctx, database); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if userID == "" && userUsername == "" {
		return fmt.Errorf("--id or --username is required")
	}

	authService := services.NewAuthService(database, auth.Config{})

	targetID := userID
	if targetID == "" {
		user, err := database.GetUserByUsername(ctx, userUsername)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		targetID = user.ID
	}

	if err := authService.DeleteUser(ctx, targetID, ""); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	fmt.Printf("✅ Deleted user %s\n", targetID)
	return nil
}
