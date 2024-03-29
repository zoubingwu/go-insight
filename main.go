package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kanmu/go-sqlfmt/sqlfmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
	Use:   "go-insight",
	Short: "Your AI-powered assistant for data analysis",
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func Run() {
	reader := bufio.NewReader(os.Stdin)

connect:
	fmt.Println("Welcome to TiInsight! Please input your database connection string:")
	fmt.Print(">> ")
	connByte, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
	connString := strings.TrimSpace(string(connByte))

	if len(connString) == 0 {
		fmt.Println("Database connection string cannot be empty")
		goto connect
	}
	parsedURL, err := url.Parse(connString)
	if err != nil {
		fmt.Println("Error parsing database connection string:", err)
		goto connect
	}

	client := NewInsightSDK()

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Testing connection for %s%s...", parsedURL.Host, parsedURL.Path)
	s.Start()

	testResult, err := client.TestConnection(connString)

	if err != nil {
		log.Fatalln(err)
	}

	if testResult.Code != 200 || testResult.Result.Pass != true {
		fmt.Println("Database connection failed, ")
		s.Disable()
		goto connect
	}

	s.Disable()
	fmt.Printf("Database %s%s connected\n", parsedURL.Host, parsedURL.Path)

	s.Suffix = " Creating context..."
	s.Enable()

	ctxResult, err := client.CreateContext(connString)
	if err != nil {
		log.Fatalln(err)
	}

	sessionResult, err := client.CreateSessionContext(ctxResult.Result.DataContextId)
	sessionId := sessionResult.Result.SessionContextId
	if err != nil {
		log.Fatalln(err)
	}
	s.Disable()

	fmt.Println("Context created, now start ask anything about your data!")

	for {
		fmt.Print(">> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == ".exit" || input == ".quit" {
			break
		}

		s.Suffix = " Thinking..."
		s.Enable()

		resp, err := client.BreakdownUserQuestion(input, sessionId)
		if err != nil {
			s.Disable()
			fmt.Println("Error occured: ", err.Error())
			continue
		}

		if resp.Result.Status == "failed" {
			s.Disable()
			fmt.Println(resp.Result.Reason)
			continue
		}

		for taskId := range resp.Result.Result.TaskTree {
			s.Suffix = " Generating and running SQL queries..."
			s.Enable()
			resp, err := client.FollowupSubTask(sessionId, taskId, resp.Result.Result.QuestionId)
			if err != nil {
				log.Fatalln(err)
			}
			s.Disable()

			sql, err := sqlfmt.Format(resp.Result.Result.Sql, &sqlfmt.Options{})
			if err != nil {
				fmt.Println("Error parsing sql: ", err.Error())
				continue
			}
			fmt.Println(strings.TrimSpace(sql))

			t := table.NewWriter()

			headers := lo.Map(resp.Result.Result.Columns, func(c Column, _ int) string {
				return c.Col
			})
			h := make(table.Row, len(headers)+1)
			h[0] = "#"
			for i, header := range headers {
				h[i+1] = header
			}

			t.AppendHeader(h)

			lo.ForEach(resp.Result.Result.Rows, func(item []any, index int) {
				row := lo.Map(item, func(item any, _ int) any {
					return fmt.Sprint(item)
				})

				r := make(table.Row, len(row)+1)
				r[0] = index + 1
				for i, a := range row {
					r[i+1] = a
				}

				t.AppendRow(r)
			})

			fmt.Println(t.Render())
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
