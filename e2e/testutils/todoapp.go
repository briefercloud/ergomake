package testutils

import (
	"fmt"

	playwright "github.com/playwright-community/playwright-go"
)

type Browser struct {
	page playwright.Page
}

func NewBrowser() (*Browser, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	chromium, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	page, err := chromium.NewPage()
	if err != nil {
		return nil, err
	}

	b := &Browser{page: page}
	return b, nil
}

func (b *Browser) LoadURL(url string) (int, error) {
	r, err := b.page.Goto(url)
	if err != nil {
		return 0, err
	}

	_, err = b.page.WaitForSelector("body")
	if err != nil {
		return 0, err
	}

	return r.Status(), nil
}

func (b *Browser) clickButton(buttonText string) error {
	buttonSelector := "//button[contains(.,'" + buttonText + "')]"

	err := b.page.Click(buttonSelector)
	if err != nil {
		return err
	}

	return nil
}

func (b *Browser) fillInput(inputName string, inputValue string) error {
	selector := fmt.Sprintf("input[name='%s'], textarea[name='%s']", inputName, inputName)

	el, err := b.page.QuerySelector(selector)
	if err != nil {
		return err
	}

	if err := el.Focus(); err != nil {
		return err
	}

	if err := el.Type(inputValue); err != nil {
		return err
	}

	return nil
}

func (b *Browser) AddTodo(todoName string) error {
	if err := b.clickButton("Add Todo"); err != nil {
		return err
	}

	if err := b.fillInput("title", todoName); err != nil {
		return err
	}

	if err := b.fillInput("description", "my todo"); err != nil {
		return err
	}

	if err := b.fillInput("duedate", "01022023"); err != nil {
		return err
	}

	if err := b.clickButton("Save"); err != nil {
		return err
	}

	return nil
}

func (b *Browser) GetTodoItems() (int, error) {
	listSelector := "ul.list-group"

	_, err := b.page.WaitForSelector(listSelector)
	if err != nil {
		return 0, err
	}

	listChildren, err := b.page.QuerySelectorAll(listSelector + " > *")
	if err != nil {
		return 0, err
	}

	count := len(listChildren)
	return count, nil
}

func (b *Browser) Refresh() error {
	_, err := b.page.Reload()
	if err != nil {
		return err
	}

	_, err = b.page.WaitForSelector("body")
	if err != nil {
		return err
	}

	return nil
}

func (b *Browser) Close() error {
	return b.page.Close()
}

// Useful for debugging
func (b *Browser) WaitSeconds(seconds int) error {
	ms := float64(seconds * 1000)
	b.page.WaitForTimeout(ms)
	return nil
}
