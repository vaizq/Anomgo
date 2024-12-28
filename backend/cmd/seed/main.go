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

const numProducts = 10
const numOrders = 100

var titles = []string{"OG kush", "Master Kush", "Skunk#1", "Gorilla Glue", "Critical", "Pure Power Plant"}
var descriptions = []string{
	"Hyvää budia",
	`Sativa-dominantti Trainwreck tuottaa vankkoja kasveja ja hyviä satoja, jos sitä kohdellaan rakkaudella ja huolenpidolla. Tämä hyvän mielen lajike miellyttää niitä, jotka nauttivat piristävästä kannabiksesta, jonka vaikutukset saavat olon energiseksi ja valmiiksi valloittamaan päivän. Herkullinen sitrushedelmien ja männyn aromi sekoittuu yrtti- ja minttuvivahteisiin tarjoten maukkaan kokemuksen.`,
	`Useimmat kasvattavat hurmaantuvat Mimosan iskevästä ulkonäöstä. Se ei kuitenkaan ole lajikkeen ainoa mainitsemisen arvoinen ominaisuus, sillä myös THC-tasot ovat tukevat ja terpeeniprofiili upea. Kompakti koko tekee Mimosasta helpon ja vaivattoman kasvatettavan. Huolimatta pienestä koostaan, onnistuu Mimosa-kannabislajike silti tuottamaan palkitsevia satoja.`,
	`Lajiketta valittaessa kasvattajalla on aina useita erilaisia vaatimuksia kasvilleen: sato, voimakkuus, maku, kasvatuksen helppous, sopivuus ilmastoon ja niin edelleen. On kuitenkin olemassa eräs kannabis lajike, joka sovittaa yhteen näitä vaatimuksia lähestyen täten täydellistä rahakasvin ideaalia.`,
	`Sisällä 600 W:n valon alla kasvatettuna nämä siemenet tuottavat kasveja, joista saadaan noin 65–75 grammaa kasvia kohti tai 600 g/m². Ulkona Critical puolestaan suosii lämmintä ilmastoa, esimerkiksi Espanjaa, Italiaa tai Kaliforniaa. Oikeissa olosuhteissa se voi tuottaa näissä ilmastoissa yli 100 grammaa kasvia kohti. Johtuen sen suhteellisen lyhyestä kukkimisajasta, voidaan sitä kasvattaa pohjoisemmilla leveysasteilla (kuten Iso-Britanniassa, Alankomaissa tai Suomessa). Siemenet eivät kuitenkaan saavuta täyttä potentiaaliaan, kuten ne sisällä tekisivät.
	Mikään muu Royal Queen Seedsin kasvi ei tuota yhtä paljon seitsemässä viikossa kuin feminisoidut Critical- kannabiksen siemenet. Lyhyen kukkimissyklin ansiosta se sopii kasvattajille, joiden täytyy viedä projekti alusta loppuun tiukan aikarajan puitteissa. Kaupalliset kasvattajat osaavat arvostaa sen ominaisuuksia Sea of Green / Screen of Green -metodeja käytettäessä, sillä se tuottaa isoja satoja vähäisellä ylläpidolla.`,
}

var deliveryDetails = []string{
	"Kontulan pirtti 1B11 Helsinki 0420",
	"Hiihtäjäntie 13 Jyväskylä 0666",
	"Lentäjänpolku 2 B 14 Iisalmi 0691",
}

var imagesFolder = "uploads/product-images"
var logoFolder = "uploads/vendor-logos"
var pricings = []product.Pricing{
	{Quantity: 1, Price: 20},
	{Quantity: 2, Price: 40},
	{Quantity: 5, Price: 90},
	{Quantity: 10, Price: 150},
	{Quantity: 100, Price: 1200},
}

var deliveryMethods = []product.DeliveryMethod{
	{Description: "Posti (kirjekurori)", Price: 4},
	{Description: "Posti (2xMylar)", Price: 6},
	{Description: "Posti (pakett)", Price: 10},
	{Description: "Maastokätkö (vain yli 100€ tilauksiin uusimaan alueella)", Price: 20},
}

var users = []string{"Pirkka", "Make", "Sakari", "BOB420"}

var reviews = []struct {
	grade   int
	message string
}{
	{grade: 5, message: "Tuote priimaa, toimitus nopea ja todella hyvä stealth pakkaus!"},
	{grade: 5, message: "Laatua niinkuin aina!"},
	{grade: 4, message: "Kukka oli hyvää ja toimitus kesti vain kolme päviää."},
	{grade: 3, message: "Ok"},
	{grade: 2, message: "Budit oli aika höttöä, muuten ok."},
	{grade: 1, message: "Tilasin 3g, mutta tulikin 2.5g ja budit ihan paskaa höttöä. Paketointi oli tehty ilmastointiteipillä. Älkää ostako!."},
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

	images := imageFilenames(imagesFolder)

	log.Println("Creating some users")
	uids := []uuid.UUID{}
	for _, user := range users {
		u, err := auth.Register(db, user, user+"123", "")
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
	}

	log.Println("Creating vendor pledges")
	logos := imageFilenames(logoFolder)
	usedLogos := map[string]bool{}
	for _, uid := range uids {
		logo := selectRandom(logos)
		for isUsed, ok := usedLogos[logo]; ok && isUsed; {
			logo = selectRandom(logos)
		}
		pledge.Create(db, uid, selectRandom(logos))
	}

	log.Println("Creating some products")
	products := make([]model.Product, 0)
	for i := 0; i < numProducts; i++ {
		// Create product
		product, err := product.Create(
			db,
			selectRandom(titles),
			selectRandom(descriptions),
			selectRandom(images),
			pricings,
			deliveryMethods,
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
