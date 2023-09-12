package app

import "C"
import (
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	services := [][]string{
		{"0", "Райнхольд Андреас Месснер", "Италия", "17 сентября 1944 - нн (78 лет)", "image/messner.png"},
		{"1", "Юзеф Е́жи «Юрек» Куку́чка", "Польша", "24 марта 1948 - 24 октября 1989 (41 год)", "image/kukuczka.png"},
		{"2", "Лоретан, Эрхард", "Швейцария", "28 апреля 1959 - 28 апреля 2011 (52 года)", "image/erhard.jpeg"},
		{"3", "Карсолио, Карлос", "Мексика", "4 октября 1962 - нн (60 лет)", "image/carsolio.jpg"},
		{"4", "Велицкий, Кшиштоф", "Польша", "5 января 1950 - нн (73 года)", "image/shishtov.jpg"},
		{"5", "Велицкий, Кшиштоф", "Польша", "5 января 1950 - нн (73 года)", "image/shishtov.jpg"},
		{"6", "Велицкий, Кшиштоф", "Польша", "5 января 1950 - нн (73 года)", "image/shishtov.jpg"},
		{"7", "Велицкий, Кшиштоф", "Польша", "5 января 1950 - нн (73 года)", "image/shishtov.jpg"},
	}

	servicesVerbose := [][]interface{}{
		{
			"0",
			"Месснер, Райнхольд",
			template.HTML("<p>Райнхольд Андреас Месснер (нем. Reinhold Andreas Messner; род. 17 сентября 1944, Бриксен) — итальянский альпинист из немецкоговорящей автономной провинции Южного Тироля, Италия, первым совершивший восхождения на все 14 «восьмитысячников» мира, некоторые из них в одиночку.</p><p> Месснер — один из самых знаменитых альпинистов в мировой истории, путешественник, писатель, в настоящее время депутат Европарламента и общественный деятель. Пионер спортивного подхода к альпинизму, ввел в практику скоростные одиночные восхождения, сначала в Доломитовых Альпах, а затем и в районе Монблана. Стал первым альпинистом, покорившим все 14 восьмитысячников мира, одним из первых достиг семи высочайших вершин континентов, двух полюсов. Совершил рекордные восхождения в разных регионах: по южной стене Аконкагуа, Брич Уолл на Килиманджаро, по юго-западной стене на Мак-Кинли и др. Но главным образом он стал известен своими восхождениями в Гималаях.</p><p>9 апреля 2010 года Райнхольду Месснеру вручён почётный Lifetime Achievement Piolet d’Or (Золотой ледоруб — самая престижная награда в альпинизме, вручаемая за выдающиеся достижения), второй в истории после Вальтера Бонатти (2009). Лауреат премии принцессы Астурийской (2018, совместно с Кшиштофом Велицким).</p>"),
			"http://localhost:8080/image/messnerBig.jpg",
			"Италия",
		},
		{
			"1",
			"Кукучка, Ежи",
			"Юзеф Е́жи «Юрек» Куку́чка (польск. Józef Jerzy Kukuczka; 24 марта 1948, Катовице, Польская Республика — 24 октября 1989, Лхоцзе, Гималаи, Непал) — польский альпинист, второй (после Райнхольда Месснера) в мире восходитель на все восьмитысячники планеты. Офицер Ордена Возрождения Польши (посмертно), золотого и серебряного Креста Заслуги, кавалер серебряного Олимпийского ордена XV зимних олимпийских игр (Калгари, 1988), одиннадцатикратный обладатель золотой медали «За выдающиеся спортивные достижения», почётный член Польской ассоциации альпинизма (PZA) (посмертно).\nДля обретения «Короны Гималаев» Ежи Кукучке понадобилось чуть меньше восьми лет (Р. Месснеру шестнадцать — на последний он поднялся 17 октября 1986 года, Кукучка менее года спустя — 18 сентября 1987-го). На десять из четырнадцати восьмитысячников планеты Ежи поднялся по новым маршрутам, на семь из них в альпийском стиле, на четыре впервые зимой. Восхождения на шесть из них были осуществлены в течение менее чем двух календарных лет.\nПогиб в результате обрыва страховочной верёвки во время попытки восхождения на Лхоцзе по Южной стене с последующим срывом на почти 2000 метров. Тело не найдено. Формально захоронено в ледниковой трещине у подножия стены.",
			"http://localhost:8080/image/kukuczkaBig.jpg",
			"Польша",
		},
		{
			"2",
			"Лоретан, Эрхард",
			"Эрхард Лоретан (англ. Erhard Loretan; 1959—2011) — швейцарский альпинист, горный гид, третий человек в мире, покоривший все 14 вершин планеты высотой более 8000 метров (второй после Райнхольда Месснера по бескислородным восхождениям).\nЭрхард Лоретан родился 28 апреля в Бюле — небольшой коммуне под Фрибургом, кантон Вале, Швейцария. По профессии плотник-краснодеревщик[2]. Альпинизмом начал заниматься в возрасте 11 лет и четыре года спустя покорил свою первую «серьёзную» вершину — Долденхорн[de] (3645 м) (по восточной стене)[3]. В 1981 году получил квалификацию горного гида[4].\n\nЕго первым восьмитысячником стала вершина Нанга-Парбат в Пакистане, которую он покорил в 1982 году. За последующие 13 лет Лоретан взошёл на все вершины выше 8000 метров, став третьим человеком в мире после Месснера и Ежи Кукучки, кому они оказались по силам. Среди его наиболее выдающихся достижений потрясающий хет-трик на Гашербрум II (8035), Гашербрум I (8068) и Броуд-Пик (8091), на которые он поднялся за 17 дней (1983), а также восхождение на Эверест в 1986 году, совершённое по северной стене по кулуару Хорнбайна за 43 часа без кислорода. Зимой этого же 1986 года Лоретан за 19 дней он покорил 38 вершин в Вальских Альпах, 30 из которых высотой более 4000 метров.\nВ 2003 году Лоретан был признан швейцарским судом виновным в непреднамеренном убийстве собственного семимесячного ребёнка и приговорён к четырём месяцам лишения свободы условно и штрафу. Несмотря на то, что обвинение настаивало на более суровом наказании, судья заявил, что известный альпинист «и так уже достаточно наказан тем, что своими руками, сам того не подозревая, погубил собственного сына»[6]. Причиной смерти младенца стал синдром детского сотрясения (СДС) — в канун Рождества 2001 года, Лоретан, пытаясь успокоить плачущего сына, слегка встряхнул его, что привело к трагическим последствиям. Альпинист отказался от анонимности процесса, и, по его словам, сам хотел, чтобы этот процесс привлёк внимание родителей к проблеме СДС и показал важность бережного отношения к маленьким детям. «Я должен жить с этой драмой до конца моих дней, до того момента, когда мы с ним встретимся вновь».\nЭрхард Лоретан погиб 28 апреля 2011 года в свой собственный день рождения во время восхождения вместе с его клиенткой Ксенией Миндер (англ. Xenia Minder) на вершину Грос-Грюнхорн в результате срыва и падения на более чем 200 метров. В своём интервью она рассказала, что потеряла равновесие на гребневом участке маршрута и сдёрнула Лоретана за собой. Сама она получила тяжёлые травмы.",
			"http://localhost:8080/image/erhard.jpeg",
			"Швейцария",
		},
		{
			"3",
			"Карсолио, Карлос",
			"Карлос Карсолио (исп. Carlos Carsolio Larrea; 1962, Мехико) — мексиканский альпинист, первый латиноамериканец, а также самый молодой восходитель (до 2002 года), покоривший все 14 восьмитысячников Земли, три из которых по новым маршрутам, а семь соло (в одиночных восхождениях) (четвёртый в общем списке и третий после Райнхольда Месснера и Эрхарда Лоретана по бескислородным восхождениям).\n\nИзвестен не только своими выдающимися альпинистскими достижениями, но и как парапланерист, режиссёр-документалист, снявший более 30 документальных фильмов как для внутримексиканского проката, так и для показа за рубежом, преподаватель и предприниматель в области популяризации альпинизма и скалолазания.",
			"http://localhost:8080/image/carsolio.jpg",
			"Мексика",
		},
		{
			"4",
			"Велицкий, Кшиштоф",
			"Кшиштоф Ежи Велицкий (польск. Krzysztof Jerzy Wielicki; 5 января 1950, Остшешув, Польша) — польский альпинист, пятый человек в мире, покоривший всё 14 восьмитысячников планеты, из которых три — Эверест, Канченджангу и Лхоцзе (соло) впервые в зимнее время года. Автор ряда рекордов восхождений на гималайские гиганты — на Броуд-Пик (первое в истории одиночное восхождение на восьмитысячник за 22 часа от основания до вершины), соло на Дхаулагири (за 16 часов) и Шишабангму по новым маршрутам.\n\nЧлен Клуба исследователей[en], обладатель одной из высших наград клуба — премии Лоуэлла Томаса[en] (англ. The Lowell Thomas Award) в номинации «альпинизм». Почётный выпускник Вроцлавского технологического университета (2007). В 2019 году удостоен самой престижной награды в мировом альпинизме «Золотой ледоруб» в номинации «За достижения всей жизни» (англ. Lifetime Achievement Award).",
			"http://localhost:8080/image/shishtov.jpg",
			"Польша",
		},
		{
			"5",
			"Велицкий, Кшиштоф",
			"Кшиштоф Ежи Велицкий (польск. Krzysztof Jerzy Wielicki; 5 января 1950, Остшешув, Польша) — польский альпинист, пятый человек в мире, покоривший всё 14 восьмитысячников планеты, из которых три — Эверест, Канченджангу и Лхоцзе (соло) впервые в зимнее время года. Автор ряда рекордов восхождений на гималайские гиганты — на Броуд-Пик (первое в истории одиночное восхождение на восьмитысячник за 22 часа от основания до вершины), соло на Дхаулагири (за 16 часов) и Шишабангму по новым маршрутам.\n\nЧлен Клуба исследователей[en], обладатель одной из высших наград клуба — премии Лоуэлла Томаса[en] (англ. The Lowell Thomas Award) в номинации «альпинизм». Почётный выпускник Вроцлавского технологического университета (2007). В 2019 году удостоен самой престижной награды в мировом альпинизме «Золотой ледоруб» в номинации «За достижения всей жизни» (англ. Lifetime Achievement Award).",
			"http://localhost:8080/image/shishtov.jpg",
			"Польша",
		},
		{
			"6",
			"Велицкий, Кшиштоф",
			"Кшиштоф Ежи Велицкий (польск. Krzysztof Jerzy Wielicki; 5 января 1950, Остшешув, Польша) — польский альпинист, пятый человек в мире, покоривший всё 14 восьмитысячников планеты, из которых три — Эверест, Канченджангу и Лхоцзе (соло) впервые в зимнее время года. Автор ряда рекордов восхождений на гималайские гиганты — на Броуд-Пик (первое в истории одиночное восхождение на восьмитысячник за 22 часа от основания до вершины), соло на Дхаулагири (за 16 часов) и Шишабангму по новым маршрутам.\n\nЧлен Клуба исследователей[en], обладатель одной из высших наград клуба — премии Лоуэлла Томаса[en] (англ. The Lowell Thomas Award) в номинации «альпинизм». Почётный выпускник Вроцлавского технологического университета (2007). В 2019 году удостоен самой престижной награды в мировом альпинизме «Золотой ледоруб» в номинации «За достижения всей жизни» (англ. Lifetime Achievement Award).",
			"http://localhost:8080/image/shishtov.jpg",
			"Польша",
		},
		{
			"7",
			"Велицкий, Кшиштоф",
			"Кшиштоф Ежи Велицкий (польск. Krzysztof Jerzy Wielicki; 5 января 1950, Остшешув, Польша) — польский альпинист, пятый человек в мире, покоривший всё 14 восьмитысячников планеты, из которых три — Эверест, Канченджангу и Лхоцзе (соло) впервые в зимнее время года. Автор ряда рекордов восхождений на гималайские гиганты — на Броуд-Пик (первое в истории одиночное восхождение на восьмитысячник за 22 часа от основания до вершины), соло на Дхаулагири (за 16 часов) и Шишабангму по новым маршрутам.\n\nЧлен Клуба исследователей[en], обладатель одной из высших наград клуба — премии Лоуэлла Томаса[en] (англ. The Lowell Thomas Award) в номинации «альпинизм». Почётный выпускник Вроцлавского технологического университета (2007). В 2019 году удостоен самой престижной награды в мировом альпинизме «Золотой ледоруб» в номинации «За достижения всей жизни» (англ. Lifetime Achievement Award).",
			"http://localhost:8080/image/shishtov.jpg",
			"Польша",
		},
	}
	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"services": services,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"services": services,
		})
	})

	router.GET("/t", func(context *gin.Context) {
		alps, err := a.repository.FilterByCountry("Польша")
		if err != nil {
			context.AbortWithStatus(404)
			return
		}

		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"services": alps,
		})
	})

	router.GET("/service/:id", func(context *gin.Context) {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.AbortWithStatus(404)
			return
		}

		if id >= len(services) || id < 0 {
			context.AbortWithStatus(404)
			return
		}

		context.HTML(http.StatusOK, "card.tmpl", gin.H{
			"servicesVerbose": servicesVerbose[id],
			"id":              id,
		})
	})

	router.GET("/filter", func(context *gin.Context) {
		searchQuery := context.DefaultQuery("name", "")
		var foundAlpinists [][]string
		for _, alpinist := range services {
			if strings.HasPrefix(strings.ToLower(alpinist[2]), strings.ToLower(searchQuery)) {
				foundAlpinists = append(foundAlpinists, alpinist)
			}
		}

		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"services": foundAlpinists,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"services": foundAlpinists,
		})
	})

	router.Static("/image", "./static/images")

	err := router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

//func (a *Application) StartServer() {
//	log.Println("Server start up")
//
//	r := gin.Default()
//
//	r.GET("/ping", func(c *gin.Context) {
//		id := c.Query("id") // получаем из запроса query string
//
//		if id != "" {
//			log.Printf("id recived %s\n", id)
//			intID, err := strconv.Atoi(id) // пытаемся привести это к чиселке
//			if err != nil {                // если не получилось
//				log.Printf("cant convert id %v", err)
//				c.Error(err)
//				return
//			}
//
//			product, err := a.repo.GetProductByID(uint(intID))
//			if err != nil { // если не получилось
//				log.Printf("cant get product by id %v", err)
//				c.Error(err)
//				return
//			}
//
//			c.JSON(http.StatusOK, gin.H{
//				"product_price": product.Price,
//			})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"message": "pong",
//		})
//	})
//
//	r.LoadHTMLGlob("templates/*")
//
//	r.GET("/test", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "test.tmpl", gin.H{
//			"title": "Main website",
//			"test":  []string{"a", "b"},
//		})
//	})
//
//	r.Static("/image", "./resources")
//
//	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
//
//	log.Println("Server down")
//}
