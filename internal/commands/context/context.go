package context

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/memberships"
	"github.com/sumup/sumup-go/shared"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/config"
	"github.com/sumup/sumup-cli/internal/display/message"
)

const (
	debounceDelay = 500 * time.Millisecond
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "context",
		Usage: "Manage merchant context for commands.",
		Commands: []*cli.Command{
			{
				Name:   "set",
				Usage:  "Set the current merchant context.",
				Action: setContext,
			},
			{
				Name:   "get",
				Usage:  "Get the current merchant context.",
				Action: getContext,
			},
			{
				Name:   "unset",
				Usage:  "Unset the current merchant context.",
				Action: unsetContext,
			},
		},
	}
}

type searchResultMsg struct {
	memberships []memberships.Membership
	err         error
}

// searchDebounceMsg is sent after the debounce delay to trigger the actual search
type searchDebounceMsg struct{}

type navigationLevel struct {
	memberships []memberships.Membership
	parentID    string
	parentType  memberships.ResourceType
	parentName  string
}

type model struct {
	ctx    context.Context
	client *sumup.Client
	// Error state
	err error
	// Stack of navigation levels for back navigation
	navigationStack []navigationLevel
	// Current level being displayed
	currentLevel navigationLevel
	// Currently displayed items (filtered or all)
	displayed []memberships.Membership
	// Current cursor position in the list
	cursor int
	// Search input field
	searchInput textinput.Model
	// Whether search mode is active
	searching bool
	// Selected membership (if any)
	selected *memberships.Membership
	// Whether a search is in progress
	loading bool
	// Last executed search query to avoid duplicate requests
	lastSearchQuery string
	// Whether a search is pending (debouncing)
	searchPending bool
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// clearSearch exits search mode and restores the current level's items
func (m *model) clearSearch() {
	m.searching = false
	m.searchInput.SetValue("")
	m.displayed = m.currentLevel.memberships
	m.lastSearchQuery = ""
	m.cursor = 0
}

// pushLevel saves current level to stack and sets a new current level
func (m *model) pushLevel(newLevel navigationLevel) {
	m.navigationStack = append(m.navigationStack, m.currentLevel)
	m.currentLevel = newLevel
	m.displayed = newLevel.memberships
	m.cursor = 0
}

// popLevel returns to the previous navigation level
func (m *model) popLevel() {
	if len(m.navigationStack) == 0 {
		return
	}
	previousLevel := m.navigationStack[len(m.navigationStack)-1]
	m.navigationStack = m.navigationStack[:len(m.navigationStack)-1]
	m.currentLevel = previousLevel
	m.displayed = previousLevel.memberships
	m.cursor = 0
}

// drillDownIntoOrg navigates into an organization to view its child merchants
func (m *model) drillDownIntoOrg(orgID, orgName string) tea.Cmd {
	parentType := memberships.ResourceType("organization")
	newLevel := navigationLevel{
		memberships: []memberships.Membership{},
		parentID:    orgID,
		parentType:  parentType,
		parentName:  orgName,
	}
	m.pushLevel(newLevel)
	m.loading = true
	return m.searchMemberships("", orgID, parentType)
}

// debounce returns a command that waits for the debounce delay before sending a searchDebounceMsg
func debounce() tea.Cmd {
	return tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
		return searchDebounceMsg{}
	})
}

// searchMemberships performs an API call to search for memberships by name
func (m model) searchMemberships(query string, parentID string, parentType memberships.ResourceType) tea.Cmd {
	return func() tea.Msg {
		status := shared.MembershipStatusAccepted
		params := memberships.ListMembershipsParams{
			Status: &status,
		}

		if query != "" {
			params.ResourceName = &query
		}

		if parentID != "" {
			params.ResourceParentId = &parentID
		}

		if parentType != "" {
			params.ResourceParentType = &parentType
		}

		response, err := m.client.Memberships.List(m.ctx, params)
		if err != nil {
			return searchResultMsg{err: err}
		}

		return searchResultMsg{memberships: response.Items}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.currentLevel.memberships = msg.memberships
		m.displayed = msg.memberships
		if m.cursor >= len(m.displayed) {
			m.cursor = max(0, len(m.displayed)-1)
		}
		return m, nil

	case searchDebounceMsg:
		if !m.searchPending {
			return m, nil
		}
		m.searchPending = false
		query := strings.TrimSpace(m.searchInput.Value())
		if query == m.lastSearchQuery {
			return m, nil
		}
		m.lastSearchQuery = query
		m.loading = true
		return m, m.searchMemberships(query, m.currentLevel.parentID, m.currentLevel.parentType)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.searching {
				m.clearSearch()
			} else if len(m.navigationStack) > 0 {
				m.popLevel()
			}
			return m, nil
		case "/":
			if !m.searching {
				m.searching = true
				return m, m.searchInput.Focus()
			}
		case "enter":
			if m.searching {
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
			if len(m.displayed) > 0 && m.cursor < len(m.displayed) {
				selectedMembership := m.displayed[m.cursor]
				if selectedMembership.Resource.Type == "organization" {
					return m, m.drillDownIntoOrg(selectedMembership.Resource.ID, selectedMembership.Resource.Name)
				} else {
					// Select merchant
					m.selected = &selectedMembership
					return m, tea.Quit
				}
			}
		case "up", "k":
			if !m.searching {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "down", "j":
			if !m.searching {
				if m.cursor < len(m.displayed)-1 {
					m.cursor++
				}
			}
		}
	}

	if m.searching {
		oldValue := m.searchInput.Value()
		m.searchInput, cmd = m.searchInput.Update(msg)
		newValue := m.searchInput.Value()

		if oldValue != newValue {
			m.searchPending = true
			m.cursor = 0
			return m, tea.Batch(cmd, debounce())
		}
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// Show title with breadcrumb if we're in an organization
	if m.currentLevel.parentName != "" {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Select a merchant from: %s", m.currentLevel.parentName)))
	} else {
		s.WriteString(titleStyle.Render("Select a merchant or organization:"))
	}
	s.WriteString("\n\n")

	if m.searching {
		s.WriteString("Search: ")
		s.WriteString(m.searchInput.View())
		if m.loading {
			s.WriteString(" (loading...)")
		}
		s.WriteString("\n\n")
	}

	items := m.displayed
	maxVisible := 10
	start := 0
	end := len(items)

	if len(items) > maxVisible {
		if m.cursor >= maxVisible/2 {
			start = m.cursor - maxVisible/2
		}
		end = start + maxVisible
		if end > len(items) {
			end = len(items)
			start = max(0, end-maxVisible)
		}
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	orgStyle := lipgloss.NewStyle().Faint(true)

	for i := start; i < end; i++ {
		membership := items[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		var line string
		if membership.Resource.Type == "organization" {
			line = fmt.Sprintf("%s Organization: %s (%s)", cursor, membership.Resource.Name, membership.Resource.ID)
			if i != m.cursor {
				line = orgStyle.Render(line)
			}
		} else {
			// Get merchant code
			code := "-"
			if membership.Resource.Attributes != nil {
				if codeVal, ok := membership.Resource.Attributes["merchant_code"].(string); ok {
					code = codeVal
				}
			}
			line = fmt.Sprintf("%s %s (%s)", cursor, membership.Resource.Name, code)
		}

		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		s.WriteString(line)
		s.WriteString("\n")
	}

	if len(items) == 0 {
		if m.loading {
			s.WriteString("Loading...")
		} else {
			s.WriteString("No items found.")
		}
		s.WriteString("\n")
	} else if len(items) > maxVisible {
		s.WriteString(fmt.Sprintf("\n(Showing %d-%d of %d)", start+1, end, len(items)))
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString("\n\n")
	if m.searching {
		s.WriteString(helpStyle.Render("esc: exit search | enter: confirm | ctrl+c/q: quit"))
	} else {
		help := "↑/↓ or j/k: navigate | /: search | enter: select"
		if len(m.navigationStack) > 0 {
			help += " | esc: back"
		}
		help += " | ctrl+c/q: quit"
		s.WriteString(helpStyle.Render(help))
	}

	return s.String()
}

func setContext(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	message.Notify("Fetching your memberships...")

	status := shared.MembershipStatusAccepted
	params := memberships.ListMembershipsParams{
		Status: &status,
	}

	response, err := appCtx.Client.Memberships.List(ctx, params)
	if err != nil {
		return fmt.Errorf("list memberships: %w", err)
	}

	if len(response.Items) == 0 {
		message.Warn("No memberships found.")
		return nil
	}

	searchInput := textinput.New()
	searchInput.Placeholder = "Type to filter..."
	searchInput.CharLimit = 50

	p := tea.NewProgram(model{
		client:          appCtx.Client,
		ctx:             ctx,
		navigationStack: []navigationLevel{},
		currentLevel: navigationLevel{
			memberships: response.Items,
		},
		displayed:   response.Items,
		searchInput: searchInput,
	})
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("run interactive selection: %w", err)
	}

	finalModel := result.(model)
	if finalModel.selected == nil {
		message.Warn("No merchant selected.")
		return nil
	}

	if finalModel.selected.Resource.Type == "organization" {
		message.Warn("Please select a merchant, not an organization.")
		return nil
	}

	// Get merchant code
	merchantCode := finalModel.selected.Resource.ID
	if merchantCode == "" {
		return fmt.Errorf("merchant code not found in membership attributes")
	}

	if err := config.SetCurrentMerchantCode(merchantCode); err != nil {
		return fmt.Errorf("save merchant context: %w", err)
	}

	message.Success("Merchant context set to: %s (%s)", finalModel.selected.Resource.Name, merchantCode)
	return nil
}

func getContext(_ context.Context, _ *cli.Command) error {
	merchantCode, err := config.GetCurrentMerchantCode()
	if err != nil {
		return fmt.Errorf("get merchant context: %w", err)
	}

	if merchantCode == "" {
		message.Notify("No merchant context set.")
		message.Notify("Use 'sumup context set' to set a merchant context.")
		return nil
	}

	message.Notify("Current merchant context: %s", merchantCode)
	return nil
}

func unsetContext(_ context.Context, _ *cli.Command) error {
	if err := config.SetCurrentMerchantCode(""); err != nil {
		return fmt.Errorf("unset merchant context: %w", err)
	}

	message.Success("Merchant context unset.")
	return nil
}
