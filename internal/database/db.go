package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // blank import: registers the SQLite driver
)

// DB is the global database connection pool shared by HTTP, gRPC, and CLI code.
var DB *sql.DB

func Init(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connection established")
	return runMigrations()
}

func runMigrations() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id            TEXT PRIMARY KEY,
		username      TEXT UNIQUE NOT NULL,
		email         TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS manga (
		id             TEXT PRIMARY KEY,
		title          TEXT NOT NULL,
		author         TEXT NOT NULL,
		genres         TEXT DEFAULT '[]',
		status         TEXT NOT NULL,
		total_chapters INTEGER DEFAULT 0,
		description    TEXT DEFAULT '',
		cover_url      TEXT DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS user_progress (
		user_id         TEXT NOT NULL,
		manga_id        TEXT NOT NULL,
		current_chapter INTEGER DEFAULT 0,
		status          TEXT NOT NULL,
		updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (manga_id) REFERENCES manga(id)
	);`

	if _, err := DB.Exec(schema); err != nil {
		return err
	}

	if _, err := DB.Exec(`ALTER TABLE manga ADD COLUMN cover_url TEXT DEFAULT ''`); err != nil {
		log.Printf("cover_url column already exists or could not be added: %v", err)
	}

	if err := seedManga(); err != nil {
		return err
	}

	log.Println("Database schema ready")
	return nil
}

func seedManga() error {
	rows := []struct {
		id, title, author, genres, status, description string
		chapters                                       int
	}{
		{"one-piece", "One Piece", "Eiichiro Oda", `["Action","Adventure","Shounen"]`, "ongoing", "A pirate adventure following Luffy and the Straw Hat crew.", 1110},
		{"naruto", "Naruto", "Masashi Kishimoto", `["Action","Adventure","Shounen"]`, "completed", "A young ninja seeks recognition and dreams of becoming Hokage.", 700},
		{"bleach", "Bleach", "Tite Kubo", `["Action","Supernatural","Shounen"]`, "completed", "Ichigo gains Soul Reaper powers and protects the living world.", 686},
		{"dragon-ball", "Dragon Ball", "Akira Toriyama", `["Action","Adventure","Shounen"]`, "completed", "Goku trains, battles, and protects Earth across legendary arcs.", 519},
		{"my-hero-academia", "My Hero Academia", "Kohei Horikoshi", `["Action","Superhero","Shounen"]`, "completed", "Students train to become heroes in a world full of quirks.", 430},
		{"demon-slayer", "Demon Slayer", "Koyoharu Gotouge", `["Action","Dark Fantasy","Shounen"]`, "completed", "Tanjiro fights demons while searching for a cure for his sister.", 205},
		{"jujutsu-kaisen", "Jujutsu Kaisen", "Gege Akutami", `["Action","Supernatural","Shounen"]`, "completed", "Sorcerers battle curses born from human negativity.", 271},
		{"black-clover", "Black Clover", "Yuki Tabata", `["Action","Fantasy","Shounen"]`, "ongoing", "Asta pursues the Wizard King title without magic.", 370},
		{"chainsaw-man", "Chainsaw Man", "Tatsuki Fujimoto", `["Action","Dark Fantasy","Shounen"]`, "ongoing", "Denji becomes Chainsaw Man and joins dangerous devil hunts.", 170},
		{"fairy-tail", "Fairy Tail", "Hiro Mashima", `["Action","Adventure","Fantasy"]`, "completed", "A wizard guild takes on magical jobs and world-level threats.", 545},
		{"sailor-moon", "Sailor Moon", "Naoko Takeuchi", `["Magical Girl","Romance","Shoujo"]`, "completed", "Usagi and her friends protect the world as Sailor Guardians.", 60},
		{"fruits-basket", "Fruits Basket", "Natsuki Takaya", `["Romance","Drama","Shoujo"]`, "completed", "Tohru becomes involved with a family cursed by zodiac spirits.", 136},
		{"ouran-high-school-host-club", "Ouran High School Host Club", "Bisco Hatori", `["Comedy","Romance","Shoujo"]`, "completed", "Haruhi joins an eccentric school host club.", 83},
		{"nana", "Nana", "Ai Yazawa", `["Drama","Romance","Josei"]`, "hiatus", "Two women named Nana chase love, music, and independence.", 84},
		{"kimi-ni-todoke", "Kimi ni Todoke", "Karuho Shiina", `["Romance","School","Shoujo"]`, "completed", "A shy girl slowly builds friendships and first love.", 123},
		{"ao-haru-ride", "Ao Haru Ride", "Io Sakisaka", `["Romance","School","Shoujo"]`, "completed", "Old feelings return when Futaba reunites with her first love.", 53},
		{"maid-sama", "Maid-sama!", "Hiro Fujiwara", `["Comedy","Romance","Shoujo"]`, "completed", "A strict student council president hides her maid cafe job.", 85},
		{"skip-beat", "Skip Beat!", "Yoshiki Nakamura", `["Comedy","Drama","Shoujo"]`, "ongoing", "Kyoko enters show business to rebuild herself.", 320},
		{"yona-of-the-dawn", "Yona of the Dawn", "Mizuho Kusanagi", `["Adventure","Fantasy","Shoujo"]`, "ongoing", "A princess gathers allies after losing her kingdom.", 260},
		{"cardcaptor-sakura", "Cardcaptor Sakura", "CLAMP", `["Magical Girl","Fantasy","Shoujo"]`, "completed", "Sakura captures magical cards released from a book.", 50},
		{"berserk", "Berserk", "Kentaro Miura", `["Dark Fantasy","Action","Seinen"]`, "ongoing", "Guts struggles through a brutal world shaped by fate and ambition.", 375},
		{"vagabond", "Vagabond", "Takehiko Inoue", `["Historical","Martial Arts","Seinen"]`, "hiatus", "A fictionalized journey of swordsman Miyamoto Musashi.", 327},
		{"vinland-saga", "Vinland Saga", "Makoto Yukimura", `["Historical","Action","Seinen"]`, "ongoing", "A Viking revenge story grows into a search for peace.", 210},
		{"monster", "Monster", "Naoki Urasawa", `["Thriller","Mystery","Seinen"]`, "completed", "A doctor hunts the consequences of saving a brilliant killer.", 162},
		{"20th-century-boys", "20th Century Boys", "Naoki Urasawa", `["Mystery","Sci-Fi","Seinen"]`, "completed", "Friends confront a conspiracy tied to their childhood stories.", 249},
		{"kingdom", "Kingdom", "Yasuhisa Hara", `["Historical","War","Seinen"]`, "ongoing", "An orphan rises through the wars of ancient China.", 790},
		{"tokyo-ghoul", "Tokyo Ghoul", "Sui Ishida", `["Horror","Action","Seinen"]`, "completed", "Kaneki is pulled into the hidden world of ghouls.", 143},
		{"parasyte", "Parasyte", "Hitoshi Iwaaki", `["Horror","Sci-Fi","Seinen"]`, "completed", "A student coexists with an alien parasite in his hand.", 64},
		{"goodnight-punpun", "Goodnight Punpun", "Inio Asano", `["Drama","Psychological","Seinen"]`, "completed", "A surreal coming-of-age story about isolation and adulthood.", 147},
		{"one-punch-man", "One-Punch Man", "ONE", `["Action","Comedy","Seinen"]`, "ongoing", "Saitama defeats any opponent with one punch and searches for meaning.", 200},
		{"chihayafuru", "Chihayafuru", "Yuki Suetsugu", `["Sports","Drama","Josei"]`, "completed", "Chihaya pursues competitive karuta with her friends.", 247},
		{"honey-and-clover", "Honey and Clover", "Chica Umino", `["Drama","Romance","Josei"]`, "completed", "Art college students navigate youth, love, and uncertainty.", 64},
		{"princess-jellyfish", "Princess Jellyfish", "Akiko Higashimura", `["Comedy","Josei","Slice of Life"]`, "completed", "A jellyfish-loving woman meets a stylish ally who changes her life.", 84},
		{"nodame-cantabile", "Nodame Cantabile", "Tomoko Ninomiya", `["Music","Comedy","Josei"]`, "completed", "Classical music students grow through rivalry and romance.", 136},
		{"wotakoi", "Wotakoi", "Fujita", `["Romance","Comedy","Josei"]`, "completed", "Adult office workers balance romance and otaku hobbies.", 86},
		{"march-comes-in-like-a-lion", "March Comes in Like a Lion", "Chica Umino", `["Drama","Slice of Life","Seinen"]`, "ongoing", "A young shogi player heals through found family.", 210},
		{"haikyuu", "Haikyu!!", "Haruichi Furudate", `["Sports","Comedy","Shounen"]`, "completed", "A short volleyball player chases the national stage.", 402},
		{"slam-dunk", "Slam Dunk", "Takehiko Inoue", `["Sports","Comedy","Shounen"]`, "completed", "A delinquent discovers basketball and team spirit.", 276},
		{"death-note", "Death Note", "Tsugumi Ohba", `["Mystery","Supernatural","Shounen"]`, "completed", "A student uses a deadly notebook and faces a genius detective.", 108},
		{"spy-x-family", "Spy x Family", "Tatsuya Endo", `["Action","Comedy","Shounen"]`, "ongoing", "A spy, assassin, and telepath form a fake family.", 100},
	}

	stmt := `INSERT OR IGNORE INTO manga
		(id, title, author, genres, status, total_chapters, description, cover_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	for _, row := range rows {
		if _, err := DB.Exec(stmt, row.id, row.title, row.author, row.genres, row.status, row.chapters, row.description, ""); err != nil {
			return err
		}
	}
	return nil
}
