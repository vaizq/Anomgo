package main

// Command line tool to easily populate data to the database

import (
	"LuomuTori/internal/config"
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/auth"
	"LuomuTori/internal/service/pledge"
	"LuomuTori/internal/service/product"
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"log"
	"math/rand"
	"os"
)

const numOrders = 40

type listing struct {
	title           string
	description     string
	image           string
	pricing         []product.Pricing
	deliveryMethods []product.DeliveryMethod
}

var listings = []listing{
	{
		title:       "Rucola",
		description: "Raikas ja mausteinen rucola, joka on kasvatettu ekologisesti ja ilman kemikaaleja. Se tuo ruokiin voimakasta makua ja toimii erinomaisena lisänä salaateissa, pastoissa ja voileivissä.",
		image:       "arugula.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 4},    // 50 g for 4 EUR
			{Quantity: 100, Price: 7},   // 100 g for 7 EUR
			{Quantity: 250, Price: 15},  // 250 g for 15 EUR
			{Quantity: 500, Price: 28},  // 500 g for 28 EUR
			{Quantity: 1000, Price: 50}, // 1 kg for 50 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Basilika",
		description: "Tämä tuore ja täyteläinen basilika on kasvatettu luonnollisesti ilman kemikaaleja. Sen aromaattinen maku tekee siitä erinomaisen lisän salaateille, pastoille ja tuoreille kastikkeille.",
		image:       "basilika.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 5},    // 50 g for 5 EUR
			{Quantity: 100, Price: 9},   // 100 g for 9 EUR
			{Quantity: 250, Price: 20},  // 250 g for 20 EUR
			{Quantity: 500, Price: 35},  // 500 g for 35 EUR
			{Quantity: 1000, Price: 60}, // 1 kg for 60 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
		},
	},
	{
		title:       "Parsakaali",
		description: "Korkealaatuinen parsakaali, joka on kasvatettu ekologisesti. Tämä vihannes on täynnä vitamiineja ja ravintoaineita, ja se on täydellinen lisä terveellisiin ruokalajeihin, kuten keittoihin ja paistoksiin.",
		image:       "broccoli.jpg",
		pricing: []product.Pricing{
			{Quantity: 100, Price: 6},   // 100 g for 6 EUR
			{Quantity: 200, Price: 11},  // 200 g for 11 EUR
			{Quantity: 500, Price: 25},  // 500 g for 25 EUR
			{Quantity: 1000, Price: 45}, // 1 kg for 45 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
		},
	},
	{
		title:       "Chili-paprika",
		description: "Tämä voimakas ja tulinen chili-paprika on kasvatettu huolella ja tarjoaa intensiivisen maun. Sen kirkas punainen väri ja tulisuus tekevät siitä erinomaisen mausteeksi kaikille ruokalajeille, joissa tarvitaan potkua.",
		image:       "chilli-pepper.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 4},    // 50 g for 4 EUR
			{Quantity: 100, Price: 7},   // 100 g for 7 EUR
			{Quantity: 250, Price: 15},  // 250 g for 15 EUR
			{Quantity: 500, Price: 28},  // 500 g for 28 EUR
			{Quantity: 1000, Price: 50}, // 1 kg for 50 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Valkosipuli",
		description: "Tuore valkosipuli, joka on käsin kerätty ja tarjoaa voimakkaan ja syvän maun. Tämä valkosipuli on täydellinen mauste keittoihin, pastoihin ja moniin muihin ruokalajeihin.",
		image:       "garlic.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 3},    // 50 g for 3 EUR
			{Quantity: 100, Price: 6},   // 100 g for 6 EUR
			{Quantity: 250, Price: 14},  // 250 g for 14 EUR
			{Quantity: 500, Price: 25},  // 500 g for 25 EUR
			{Quantity: 1000, Price: 45}, // 1 kg for 45 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Limetti",
		description: "Raikas ja mehevä limetti, joka tuo eloisuutta ruokiin ja juomiin. Tämä ekologisesti kasvatettu limetti on täydellinen makuvivahde salaateille, juomille ja ruoille.",
		image:       "lime.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 4},    // 50 g for 4 EUR
			{Quantity: 100, Price: 8},   // 100 g for 8 EUR
			{Quantity: 250, Price: 18},  // 250 g for 18 EUR
			{Quantity: 500, Price: 30},  // 500 g for 30 EUR
			{Quantity: 1000, Price: 50}, // 1 kg for 50 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Paprika",
		description: "Värikäs ja mehukas paprika, joka on täynnä vitamiineja ja raikkautta. Tämä eksoottinen paprika on kasvatettu parhaille kentille ja se on täydellinen lisä moniin ruokalajeihin.",
		image:       "paprika.jpg",
		pricing: []product.Pricing{
			{Quantity: 100, Price: 5},   // 100 g for 5 EUR
			{Quantity: 200, Price: 9},   // 200 g for 9 EUR
			{Quantity: 500, Price: 18},  // 500 g for 18 EUR
			{Quantity: 1000, Price: 35}, // 1 kg for 35 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Pinaatti",
		description: "Tuore pinaatti, joka on täynnä ravinteita ja makua. Tämä monipuolinen vihannes sopii keittoihin, salaateihin ja smoothieihin.",
		image:       "spinach.jpg",
		pricing: []product.Pricing{
			{Quantity: 50, Price: 3},    // 50 g for 3 EUR
			{Quantity: 100, Price: 6},   // 100 g for 6 EUR
			{Quantity: 250, Price: 12},  // 250 g for 12 EUR
			{Quantity: 500, Price: 20},  // 500 g for 20 EUR
			{Quantity: 1000, Price: 35}, // 1 kg for 35 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
	{
		title:       "Tomaatit",
		description: "Tuoreet ja mehukkaat tomaatit, jotka ovat täynnä makua. Ne ovat kasvatettu korkealaatuisilla menetelmillä ja ovat täydellisiä salaateihin, pastoihin ja kaikenlaisiin ruokiin.",
		image:       "tomatos.jpg",
		pricing: []product.Pricing{
			{Quantity: 100, Price: 4},   // 100 g for 4 EUR
			{Quantity: 200, Price: 7},   // 200 g for 7 EUR
			{Quantity: 500, Price: 16},  // 500 g for 16 EUR
			{Quantity: 1000, Price: 30}, // 1 kg for 30 EUR
		},
		deliveryMethods: []product.DeliveryMethod{
			{Description: "Posti (kirje)", Price: 4},
			{Description: "Posti (paketti)", Price: 8},
			{Description: "Kotiin kuljetus", Price: 12},
			{Description: "Drone kuljetus", Price: 25}, // Added Drone delivery
		},
	},
}

type vendor struct {
	name string
	logo string
}

var vendors = []vendor{
	{name: "burgerking", logo: "bk.png"},
	{name: "cocacola", logo: "cocacola.png"},
	{name: "mercedes", logo: "mercedes.png"},
	{name: "tesla", logo: "tesla.png"},
}

var deliveryDetails = []string{
	"Aleksanterinkatu 20 A 6, Helsinki 00100",
	"Pasilanraitio 2 B 15, Helsinki 00240",
	"Rautatienkatu 10 C, Tampere 33100",
	"Kallionkatu 23 D 4, Turku 20100",
	"Rinnepolku 5, Oulu 90250",
}

var reviews = []struct {
	grade   int
	message string
}{
	{grade: 5, message: "Erittäin herkullista ja tuoretta sekä nopea toimitus!"},
	{grade: 5, message: "Taattua laatua!"},
	{grade: 4, message: "Hyvää oli ja toimitus kesti vain kolme päviää."},
	{grade: 3, message: "Kaikki ok"},
	{grade: 2, message: "Toimitus kesti tosi kauan"},
	{grade: 1, message: "Tilasin 2 viikkoa sitten eikä ole vieläkään tullut perille! Mikä mättää??"},
}

func selectRandom[T any](s []T) T {
	return s[rand.Intn(len(s))]
}

func imageFilenames(dir string) []string {
	de, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	files := []string{}
	for _, entry := range de {
		files = append(files, entry.Name())
	}
	return files
}

func main() {
	config.Parse()
	db, err := openDB(config.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var uids = []uuid.UUID{}
	log.Println("Creating some users")
	for _, vendor := range vendors {
		u, err := auth.Register(db, vendor.name, vendor.name+"123", "")
		if err != nil {
			log.Fatal(err)
		}

		uids = append(uids, u.ID)

		wallet, err := model.M.Wallet.GetForUser(db, u.ID)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := model.M.Wallet.AddBalance(db, wallet.ID, 10000000000000); err != nil {
			log.Fatal(err)
		}

		pledge.Create(db, u.ID, vendor.logo)
	}

	log.Println("Creating some products")
	products := make([]model.Product, 0)
	for _, listing := range listings {
		// Create product
		product, err := product.Create(
			db,
			listing.title,
			listing.description,
			listing.image,
			listing.pricing,
			listing.deliveryMethods,
			selectRandom(uids))
		if err != nil {
			log.Fatal(err)
		}

		products = append(products, *product)
	}

	log.Println("Creating some orders and reviews")
	for i := 0; i < numOrders; i++ {
		product := selectRandom(products)
		prices, err := model.M.Price.GetAll(db, product.ID)
		if err != nil {
			log.Fatal(err)
		}
		price := selectRandom(prices)

		dms, err := model.M.DeliveryMethod.GetAllForProduct(db, product.ID)
		if err != nil {
			log.Fatal(err)
		}
		dm := selectRandom(dms)

		customerID := product.VendorID
		for customerID == product.VendorID {
			customerID = selectRandom(uids)
		}

		order, err := model.M.Order.Create(db, price.ID, dm.ID, customerID, model.StatusCompleted, selectRandom(deliveryDetails))
		if err != nil {
			log.Fatal(err)
		}

		if _, err := model.M.DeliveryInfo.Create(db, "Lähetystunnus: 0x12f72ad33069420", order.ID); err != nil {
			log.Fatal(err)
		}

		review := selectRandom(reviews)
		if _, err := model.M.Review.Create(db, review.grade, review.message, order.ID); err != nil {
			log.Fatal(err)
		}
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
