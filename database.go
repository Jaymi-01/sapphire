package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := seedInitialData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	slog.Info("Database initialized successfully", "path", dbPath)
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS scenes (
		id TEXT PRIMARY KEY,
		location TEXT,
		title TEXT,
		body TEXT
	);

	CREATE TABLE IF NOT EXISTS choices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scene_id TEXT,
		label TEXT,
		next_scene_id TEXT,
		req_var TEXT,
		effect_var TEXT,
		effect_item TEXT,
		effect_ability TEXT,
		effect_flag TEXT,
		effect_trait TEXT,
		FOREIGN KEY(scene_id) REFERENCES scenes(id)
	);

	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE,
		type TEXT,
		power INTEGER,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS inventory (
		item_name TEXT PRIMARY KEY,
		quantity INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS player_variables (
		var_name TEXT PRIMARY KEY,
		var_value INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS player_abilities (
		ability_name TEXT PRIMARY KEY,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS player_traits (
		trait_name TEXT PRIMARY KEY,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS player_status_effects (
		effect_name TEXT PRIMARY KEY,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS world_flags (
		flag_name TEXT PRIMARY KEY,
		is_active BOOLEAN DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS player_meta (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	`
	_, err := DB.Exec(schema)
	return err
}

func seedInitialData() error {
	// Wipe for full sync
	DB.Exec("DELETE FROM choices")
	DB.Exec("DELETE FROM scenes")
	DB.Exec("DELETE FROM items")
	DB.Exec("DELETE FROM inventory")
	DB.Exec("DELETE FROM player_variables")
	DB.Exec("DELETE FROM player_abilities")
	DB.Exec("DELETE FROM player_traits")
	DB.Exec("DELETE FROM player_status_effects")
	DB.Exec("DELETE FROM world_flags")

	// 1. Magical Relics
	items := [][]interface{}{
		{"Vaelen's Cylinder", "relic", 100, "A conduit containing the names of the Unbound Souls."},
		{"Void-Staff", "catalyst", 250, "Channels the vacuum between stars. Siphons magic from reality."},
		{"Glass Focus", "catalyst", 50, "Elara's tool for weaving ley-line energy."},
	}
	for _, itm := range items {
		DB.Exec("INSERT OR IGNORE INTO items (name, type, power, description) VALUES (?, ?, ?, ?)", itm...)
	}

	// 2. The Arcane Epic
	scenes := []struct{id, loc, title, body string}{
		{"1.0", "The High Border Keep", "The Aetheric Breach", "The sky is a swirling vortex of violet light. You wake in the Nave, where Commander Vaelen is slowly being encased in a chrysalis of amethyst—the Stone-Sickness. He hands you a Cylinder etched with glowing runes. 'The High Somnomancer has traded our souls for silence. Take this to the Seer in the Basin.'\n\nTo your right, the Runic Thralls advance with blades of solidified moonlight."},
		{"1.1", "The Whisperwoods", "The Weaver of Glass", "You sprint into the woods, the Cylinder thrumming against your heart. A girl with milky, sightless eyes blocks your path. She holds a dagger of enchanted glass. 'The Cylinder is the Aetheric Census,' she whispers. 'I am Elara, apprentice to the Seer. Give it to me, or the Pulse-Oaks will weave your soul into their roots.'"},
		{"2.4", "The Weeping Basin", "The Great Dispelling", "You stand amidst the shattered remains of three Gilded Colossi. You feel a surge of raw Aether-Scarring. Your skin has turned to translucent quartz, revealing the violet fire beneath. You have erased the Chancellor's elite Coven. The path to the Capital is now a jagged scar across the ley-lines."},
		{"2.5", "The Ley-Stream", "The Journey of Will", "You lead your army of Unbound Souls to the Ley-Stream—a river of pure magical frequency. You stand at the edge, your quartz-arm glowing. To enter the stream is to risk being torn apart by the very fabric of Althoria."},
		{"3.0", "The Palace of Dreams", "The Final Conjuration", "You burst through the Palace Gates. The Chancellor sits upon a throne of petrified stars. He looks at your quartz-arm and smiles sadly. 'You have become the perfect vessel for the Ending. Will you seal the stars away, or will you let the feast begin?'"},
	}
	for _, s := range scenes {
		DB.Exec("INSERT OR REPLACE INTO scenes (id, location, title, body) VALUES (?, ?, ?, ?)", s.id, s.loc, s.title, s.body)
	}

	// 3. Choices
	choices := []struct{sID, label, next, vEff, aEff string}{
		{"1.0", "Take the Cylinder and weave a shield.", "1.1", "resonance + 10", ""},
		{"1.1", "Offer the Cylinder to Elara.", "2.4", "elara_trust + 10", "Veil-Sight"},
		{"2.4", "Siphon the remaining mana from the Colossi.", "2.5", "aether_scarring + 15", "Soul-Devourer"},
		{"2.5", "Super-Overclock the Ley-Stream with your own soul.", "3.0", "aether_scarring + 10", ""},
	}
	for _, c := range choices {
		DB.Exec("INSERT OR REPLACE INTO choices (scene_id, label, next_scene_id, effect_var, effect_ability) VALUES (?, ?, ?, ?, ?)",
			c.sID, c.label, c.next, c.vEff, c.aEff)
	}

	// 4. Initial Variables
	vars := []string{"resonance", "aether_scarring", "elara_trust", "willpower", "mana"}
	for _, v := range vars {
		DB.Exec("INSERT OR IGNORE INTO player_variables (var_name, var_value) VALUES (?, 0)", v)
	}
	DB.Exec("UPDATE player_variables SET var_value = 100 WHERE var_name = 'mana'")

	DB.Exec("INSERT OR IGNORE INTO player_meta (key, value) VALUES ('current_scene', '1.0')")

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
