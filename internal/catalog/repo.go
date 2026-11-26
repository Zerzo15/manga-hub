package catalog

import (
	"database/sql"
	"fmt"
)

type Manga struct {
	ID            int64
	Title         string
	Author        string
	Genres        string
	Status        string
	TotalChapters int
	Description   string
}

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepo(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS manga (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		genres TEXT NOT NULL,
		status TEXT NOT NULL,
		total_chapters INTEGER NOT NULL DEFAULT 0,
		description TEXT
	);`
	_, err := r.db.Exec(schema)
	return err
}

func (r *CatalogRepository) SeedIfEmpty() error {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		fmt.Printf("Catalog already seeded (%d manga)\n", count)
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO manga (title, author, genres, status, total_chapters, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range seedData {
		if _, err := stmt.Exec(m.Title, m.Author, m.Genres, m.Status, m.TotalChapters, m.Description); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("Seeded %d manga records\n", len(seedData))
	return nil
}

var seedData = []Manga{
	{Title: "One Piece", Author: "Eiichiro Oda", Genres: "Action,Adventure", Status: "Ongoing", TotalChapters: 1120, Description: "A pirate crew sails the Grand Line in search of the One Piece."},
	{Title: "Naruto", Author: "Masashi Kishimoto", Genres: "Action,Adventure", Status: "Completed", TotalChapters: 700, Description: "A young ninja dreams of becoming Hokage."},
	{Title: "Attack on Titan", Author: "Hajime Isayama", Genres: "Action,Dark Fantasy", Status: "Completed", TotalChapters: 139, Description: "Humanity fights for survival behind giant walls against Titans."},
	{Title: "Berserk", Author: "Kentaro Miura", Genres: "Action,Dark Fantasy", Status: "Ongoing", TotalChapters: 376, Description: "A lone mercenary battles demons in a brutal medieval world."},
	{Title: "Fullmetal Alchemist", Author: "Hiromu Arakawa", Genres: "Action,Fantasy", Status: "Completed", TotalChapters: 108, Description: "Two brothers search for the Philosopher's Stone to restore their bodies."},
	{Title: "Death Note", Author: "Tsugumi Ohba", Genres: "Mystery,Thriller", Status: "Completed", TotalChapters: 108, Description: "A student finds a notebook that kills anyone whose name is written in it."},
	{Title: "Monster", Author: "Naoki Urasawa", Genres: "Mystery,Thriller", Status: "Completed", TotalChapters: 162, Description: "A surgeon tracks down a former patient who became a serial killer."},
	{Title: "Detective Conan", Author: "Gosho Aoyama", Genres: "Mystery,Comedy", Status: "Ongoing", TotalChapters: 1130, Description: "A teen detective shrunk into a child solves crimes."},
	{Title: "20th Century Boys", Author: "Naoki Urasawa", Genres: "Mystery,Sci-Fi", Status: "Completed", TotalChapters: 249, Description: "Childhood friends uncover a cult using their old doomsday story."},
	{Title: "Steins;Gate", Author: "Yomi Sarachi", Genres: "Sci-Fi,Thriller", Status: "Completed", TotalChapters: 23, Description: "A self-proclaimed mad scientist accidentally invents time travel."},
	{Title: "Akira", Author: "Katsuhiro Otomo", Genres: "Sci-Fi,Action", Status: "Completed", TotalChapters: 120, Description: "A biker gang member develops dangerous psychic powers in Neo-Tokyo."},
	{Title: "Pluto", Author: "Naoki Urasawa", Genres: "Sci-Fi,Mystery", Status: "Completed", TotalChapters: 65, Description: "A robot detective investigates the destruction of the world's strongest robots."},
	{Title: "Planetes", Author: "Makoto Yukimura", Genres: "Sci-Fi,Slice of Life", Status: "Completed", TotalChapters: 26, Description: "Astronauts collect dangerous space debris orbiting Earth."},
	{Title: "Vagabond", Author: "Takehiko Inoue", Genres: "Action,Historical", Status: "Hiatus", TotalChapters: 327, Description: "A fictionalized journey of legendary swordsman Miyamoto Musashi."},
	{Title: "Vinland Saga", Author: "Makoto Yukimura", Genres: "Action,Historical", Status: "Ongoing", TotalChapters: 215, Description: "A young Viking seeks revenge before searching for a peaceful land."},
	{Title: "Kingdom", Author: "Yasuhisa Hara", Genres: "Action,Historical", Status: "Ongoing", TotalChapters: 810, Description: "A war orphan aims to become the greatest general under the heavens."},
	{Title: "Slam Dunk", Author: "Takehiko Inoue", Genres: "Sports,Comedy", Status: "Completed", TotalChapters: 276, Description: "A delinquent joins his high school basketball team to impress a girl."},
	{Title: "Haikyuu!!", Author: "Haruichi Furudate", Genres: "Sports,Drama", Status: "Completed", TotalChapters: 402, Description: "A short volleyball player strives to reach the national championship."},
	{Title: "Hajime no Ippo", Author: "George Morikawa", Genres: "Sports,Action", Status: "Ongoing", TotalChapters: 1470, Description: "A shy high schooler discovers his talent for boxing."},
	{Title: "Blue Lock", Author: "Muneyuki Kaneshiro", Genres: "Sports,Thriller", Status: "Ongoing", TotalChapters: 280, Description: "300 strikers compete in a ruthless program to create Japan's best forward."},
	{Title: "Kaguya-sama: Love Is War", Author: "Aka Akasaka", Genres: "Romance,Comedy", Status: "Completed", TotalChapters: 281, Description: "Two proud student council elites try to make the other confess first."},
	{Title: "Horimiya", Author: "HERO", Genres: "Romance,Slice of Life", Status: "Completed", TotalChapters: 122, Description: "Two classmates discover each other's hidden sides outside school."},
	{Title: "Fruits Basket", Author: "Natsuki Takaya", Genres: "Romance,Fantasy", Status: "Completed", TotalChapters: 136, Description: "An orphan girl lives with a family cursed by the Chinese Zodiac."},
	{Title: "Your Lie in April", Author: "Naoshi Arakawa", Genres: "Romance,Drama", Status: "Completed", TotalChapters: 44, Description: "A former piano prodigy rediscovers music through a free-spirited violinist."},
	{Title: "Yotsuba&!", Author: "Kiyohiko Azuma", Genres: "Slice of Life,Comedy", Status: "Ongoing", TotalChapters: 112, Description: "An energetic five-year-old explores everyday life in a new town."},
	{Title: "Barakamon", Author: "Satsuki Yoshino", Genres: "Slice of Life,Comedy", Status: "Completed", TotalChapters: 139, Description: "A calligrapher moves to a rural island to find new inspiration."},
	{Title: "Frieren: Beyond Journey's End", Author: "Kanehito Yamada", Genres: "Fantasy,Adventure", Status: "Ongoing", TotalChapters: 135, Description: "An immortal elf reflects on mortality after her party defeats the Demon King."},
	{Title: "Chainsaw Man", Author: "Tatsuki Fujimoto", Genres: "Action,Dark Fantasy", Status: "Ongoing", TotalChapters: 180, Description: "A debt-ridden teen merges with his chainsaw devil dog to hunt devils."},
	{Title: "Jujutsu Kaisen", Author: "Gege Akutami", Genres: "Action,Fantasy", Status: "Completed", TotalChapters: 271, Description: "A high schooler swallows a cursed finger and joins jujutsu sorcerers."},
	{Title: "Spy x Family", Author: "Tatsuya Endo", Genres: "Comedy,Action", Status: "Ongoing", TotalChapters: 105, Description: "A spy, an assassin, and a telepath form a fake family for a secret mission."},
}
