package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Choice struct {
	ID            int
	Label         string
	NextSceneID   string
	ReqVar        string
	EffectVar     string
	EffectItem    string
	EffectAbility string
	EffectFlag    string
	EffectTrait   string
	IsLocked      bool
	LockReason    string
}

type Scene struct {
	ID       string
	Location string
	Title    string
	Body     string
	Choices  []Choice
}

type Model struct {
	currentScene    Scene
	playerVariables map[string]int
	inventory       []string
	abilities       []string
	traits          []string
	statusEffects   []string
	worldFlags      []string
	cursor          int
	width, height   int
}

func NewModel() Model {
	return Model{
		playerVariables: make(map[string]int),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.currentScene.Choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.currentScene.Choices) > 0 {
				c := m.currentScene.Choices[m.cursor]
				if !c.IsLocked {
					return m.transition(c)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Header
	header := titleStyle.Render(fmt.Sprintf(" SAPPHIRE: %s ", strings.ToUpper(m.currentScene.Title)))
	loc := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("📍 " + m.currentScene.Location)
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, header, "  ", loc) + "\n\n")

	// Narrative Body
	bodyStyle := lipgloss.NewStyle().
		Width(85).
		Padding(1, 3).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63"))
	s.WriteString(bodyStyle.Render(m.currentScene.Body) + "\n\n")

	// Choices
	s.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).Render("THE CONSCIOUS CHOICE") + "\n")
	for i, c := range m.currentScene.Choices {
		cursor := "  "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursor = "❯ "
			style = style.Foreground(lipgloss.Color("205")).Bold(true)
		}
		
		label := c.Label
		if c.IsLocked {
			style = style.Foreground(lipgloss.Color("240")).Strikethrough(true)
			label += fmt.Sprintf(" (Locked: %s)", c.LockReason)
		}
		
		s.WriteString(style.Render(fmt.Sprintf("%s%d. %s", cursor, i+1, label)) + "\n")
	}

	// Status Bar
	s.WriteString("\n" + m.renderStatusBar())

	return s.String()
}

func (m Model) renderStatusBar() string {
	var sections []string
	
	// Core Magic Vars
	vParts := []string{}
	important := []string{"resonance", "aether_scarring", "mana"}
	for _, k := range important {
		vParts = append(vParts, fmt.Sprintf("%s: %d", strings.ToUpper(k), m.playerVariables[k]))
	}
	sections = append(sections, strings.Join(vParts, " | "))

	// Powers
	if len(m.abilities) > 0 {
		sections = append(sections, lipgloss.NewStyle().Foreground(blue).Render("SPELLS: "+strings.Join(m.abilities, ", ")))
	}

	// Artifacts
	sections = append(sections, fmt.Sprintf("RELICS: %d", len(m.inventory)))

	bar := strings.Join(sections, "  ║  ")
	return lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("250")).
		Padding(0, 2).
		Width(95).
		Render(bar)
}

func (m *Model) transition(c Choice) (tea.Model, tea.Cmd) {
	m.applyEffects(c.EffectVar, c.EffectItem, c.EffectAbility, c.EffectFlag, c.EffectTrait)
	m.SaveGameState(c.NextSceneID)
	m.LoadScene(c.NextSceneID)
	m.cursor = 0
	return m, nil
}

func (m *Model) applyEffects(varStr, itemStr, abilityStr, flagStr, traitStr string) {
	if varStr != "" {
		re := regexp.MustCompile(`(\w+)\s*([+-])\s*(\d+)`)
		matches := re.FindAllStringSubmatch(varStr, -1)
		for _, match := range matches {
			key, op := match[1], match[2]
			val, _ := strconv.Atoi(match[3])
			if op == "+" { m.playerVariables[key] += val } else { m.playerVariables[key] -= val }
			DB.Exec("INSERT INTO player_variables (var_name, var_value) VALUES (?, ?) ON CONFLICT(var_name) DO UPDATE SET var_value = ?", key, m.playerVariables[key], m.playerVariables[key])
		}
	}

	if itemStr != "" {
		if strings.HasPrefix(itemStr, "+ ") {
			itemName := strings.TrimPrefix(itemStr, "+ ")
			DB.Exec("INSERT INTO inventory (item_name, quantity) VALUES (?, 1) ON CONFLICT(item_name) DO UPDATE SET quantity = quantity + 1", itemName)
		}
	}

	if abilityStr != "" {
		DB.Exec("INSERT OR IGNORE INTO player_abilities (ability_name, description) VALUES (?, ?)", abilityStr, "Unlocked.")
	}

	if flagStr != "" {
		DB.Exec("INSERT OR REPLACE INTO world_flags (flag_name, is_active) VALUES (?, 1)", flagStr)
	}

	if traitStr != "" {
		DB.Exec("INSERT OR IGNORE INTO player_traits (trait_name, description) VALUES (?, ?)", traitStr, "Passive effect.")
	}
}

func (m *Model) LoadScene(id string) {
	err := DB.QueryRow("SELECT id, location, title, body FROM scenes WHERE id = ?", id).Scan(&m.currentScene.ID, &m.currentScene.Location, &m.currentScene.Title, &m.currentScene.Body)
	if err != nil {
		return
	}

	rows, err := DB.Query("SELECT id, label, next_scene_id, req_var, effect_var, effect_item, effect_ability, effect_flag, effect_trait FROM choices WHERE scene_id = ?", id)
	if err != nil {
		return
	}
	defer rows.Close()

	m.currentScene.Choices = []Choice{}
	for rows.Next() {
		var c Choice
		rows.Scan(&c.ID, &c.Label, &c.NextSceneID, &c.ReqVar, &c.EffectVar, &c.EffectItem, &c.EffectAbility, &c.EffectFlag, &c.EffectTrait)
		locked, reason := m.checkRequirements(c.ReqVar)
		c.IsLocked = locked
		c.LockReason = reason
		m.currentScene.Choices = append(m.currentScene.Choices, c)
	}
}

func (m Model) checkRequirements(reqStr string) (bool, string) {
	if reqStr == "" { return false, "" }
	re := regexp.MustCompile(`(\w+)\s*(>=|<=|>|<|==)\s*(\d+)`)
	match := re.FindStringSubmatch(reqStr)
	if len(match) == 4 {
		key, op := match[1], match[2]
		target, _ := strconv.Atoi(match[3])
		current := m.playerVariables[key]
		met := false
		switch op {
		case ">=": met = current >= target
		case "<=": met = current <= target
		case ">":  met = current > target
		case "<":  met = current < target
		case "==": met = current == target
		}
		if !met { return true, fmt.Sprintf("Requires %s %s %d", key, op, target) }
	}
	return false, ""
}

func (m *Model) LoadGameState() error {
	// Vars
	rows, err := DB.Query("SELECT var_name, var_value FROM player_variables")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var v int
		rows.Scan(&k, &v)
		m.playerVariables[k] = v
	}

	// Abilities
	aRows, err := DB.Query("SELECT ability_name FROM player_abilities")
	if err == nil {
		defer aRows.Close()
		m.abilities = []string{}
		for aRows.Next() { var n string; aRows.Scan(&n); m.abilities = append(m.abilities, n) }
	}

	// Traits
	tRows, err := DB.Query("SELECT trait_name FROM player_traits")
	if err == nil {
		defer tRows.Close()
		m.traits = []string{}
		for tRows.Next() { var n string; tRows.Scan(&n); m.traits = append(m.traits, n) }
	}

	// Status
	sRows, err := DB.Query("SELECT effect_name FROM player_status_effects")
	if err == nil {
		defer sRows.Close()
		m.statusEffects = []string{}
		for sRows.Next() { var n string; sRows.Scan(&n); m.statusEffects = append(m.statusEffects, n) }
	}

	// Flags
	fRows, err := DB.Query("SELECT flag_name FROM world_flags WHERE is_active = 1")
	if err == nil {
		defer fRows.Close()
		m.worldFlags = []string{}
		for fRows.Next() { var n string; fRows.Scan(&n); m.worldFlags = append(m.worldFlags, n) }
	}

	// Meta
	var sceneID string
	err = DB.QueryRow("SELECT value FROM player_meta WHERE key = 'current_scene'").Scan(&sceneID)
	if err != nil { sceneID = "1.0" }
	m.LoadScene(sceneID)
	return nil
}

func (m *Model) SaveGameState(sceneID string) {
	DB.Exec("UPDATE player_meta SET value = ? WHERE key = 'current_scene'", sceneID)
}
