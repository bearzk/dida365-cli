package cmd

import (
	"fmt"

	"github.com/bearzk/dida365-cli/internal/models"
	"github.com/spf13/cobra"
)

// Global variables for task flags
var (
	taskTitle     string
	taskProjectID string
	taskContent   string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, read, update, delete, and complete tasks.`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task in a specific project.`,
	RunE:  runTaskCreate,
}

var taskGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get task details",
	Long:  `Retrieve detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskGet,
}

var taskListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List all tasks in a project",
	Long:  `Retrieve and display all tasks in a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskList,
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update a task",
	Long:  `Update an existing task's title and/or content.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskUpdate,
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <task-id>",
	Short: "Mark a task as complete",
	Long:  `Mark a specific task as complete.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskComplete,
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Long:  `Delete a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDelete,
}

func init() {
	// Add task command to root
	rootCmd.AddCommand(taskCmd)

	// Add subcommands to task
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	// Flags for create command
	taskCreateCmd.Flags().StringVar(&taskTitle, "title", "", "Task title (required)")
	taskCreateCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskCreateCmd.Flags().StringVar(&taskContent, "content", "", "Task content (optional)")
	taskCreateCmd.MarkFlagRequired("title")
	taskCreateCmd.MarkFlagRequired("project-id")

	// Flags for get command
	taskGetCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskGetCmd.MarkFlagRequired("project-id")

	// Flags for update command
	taskUpdateCmd.Flags().StringVar(&taskTitle, "title", "", "Task title (optional)")
	taskUpdateCmd.Flags().StringVar(&taskContent, "content", "", "Task content (optional)")
	taskUpdateCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskUpdateCmd.MarkFlagRequired("project-id")

	// Flags for complete command
	taskCompleteCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskCompleteCmd.MarkFlagRequired("project-id")

	// Flags for delete command
	taskDeleteCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskDeleteCmd.MarkFlagRequired("project-id")
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	c := loadClient()

	taskCreate := &models.TaskCreate{
		Title:     taskTitle,
		ProjectID: taskProjectID,
		Content:   taskContent,
	}

	task, err := c.CreateTask(taskCreate)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(task)
	return nil
}

func runTaskGet(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	c := loadClient()

	task, err := c.GetTask(taskProjectID, taskID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(task)
	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	tasks, err := c.ListTasks(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(tasks)
	return nil
}

func runTaskUpdate(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	c := loadClient()

	// Build update payload with only changed fields
	updates := &models.TaskUpdate{}
	hasChanges := false

	if cmd.Flags().Changed("title") {
		updates.Title = &taskTitle
		hasChanges = true
	}

	if cmd.Flags().Changed("content") {
		updates.Content = &taskContent
		hasChanges = true
	}

	if !hasChanges {
		outputError(fmt.Errorf("no fields to update"), "VALIDATION_ERROR", 1)
		return nil
	}

	task, err := c.UpdateTask(taskID, updates)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(task)
	return nil
}

func runTaskComplete(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	c := loadClient()

	err := c.CompleteTask(taskProjectID, taskID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(map[string]string{
		"status":  "completed",
		"task_id": taskID,
	})
	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	c := loadClient()

	err := c.DeleteTask(taskProjectID, taskID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(map[string]string{
		"status":  "deleted",
		"task_id": taskID,
	})
	return nil
}
