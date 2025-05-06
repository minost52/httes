// package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"io/ioutil"
// 	"net/url"
// 	"os"
// 	"os/signal"
// 	"regexp"
// 	"strings"

// 	"go.ddosify.com/ddosify/config"
// 	"go.ddosify.com/ddosify/core"
// 	"go.ddosify.com/ddosify/core/proxy"
// 	"go.ddosify.com/ddosify/core/types"
// )

// const headerRegexp = `^*(.+):\s*(.+)`

// var (
// 	iterCount = flag.Int("n", types.DefaultIterCount, "Общее количество итераций")
// 	duration  = flag.Int("d", types.DefaultDuration, "Продолжительность теста в секундах")
// 	loadType  = flag.String("l", types.DefaultLoadType, "Тип нагрузочного теста [линейный, инкрементальный, волнообразный]")

// 	method = flag.String("m", types.DefaultMethod,
// 		"Введите метод запроса. Для Http(ов):[GET, POST, PUT, DELETE, UPDATE, PATCH]")
// 	payload = flag.String("b", "", "Полезная нагрузка сетевого пакета (тела)")
// 	auth    = flag.String("a", "", "Обычная аутентификация, имя пользователя:пароль")
// 	headers header

// 	target  = flag.String("t", "", "Целевой URL-адрес для отправки теста")
// 	timeout = flag.Int("T", types.DefaultTimeout, "Время ожидания запроса в секундах")

// 	proxyFlag = flag.String("P", "",
// 		"Адрес прокси-сервера в виде protocol://username:password@host:port. Поддерживаемые прокси-серверы [http(s), socks]")
// 	output = flag.String("o", types.DefaultOutputType, "Output destination")

// 	configPath = flag.String("config", "",
// 		"Путь к конфигурационному файлу в формате Json. Если указан конфигурационный файл, другие значения флага будут проигнорированы")

// 	certPath    = flag.String("cert_path", "", "Путь к файлу сертификата (обычно называемому 'cert.pem')")
// 	certKeyPath = flag.String("cert_key_path", "", "Путь к файлу ключа сертификата (обычно называемому 'key.pem')")

// 	debug = flag.Bool("debug", false, "Повторяет сценарий один раз и выводит подробный результат в виде curl")
// )

// func main() {
// 	flag.Var(&headers, "h", "Заголовки запросов. Например: -h 'Принять: text/html' -h 'Тип содержимого: application/xml'")
// 	flag.Parse()

// 	h, err := createHeart()

// 	if err != nil {
// 		exitWithMsg(err.Error())
// 	}

// 	if err := h.Validate(); err != nil {
// 		exitWithMsg(err.Error())
// 	}

// 	run(h)
// }

// func createHeart() (h types.Heart, err error) {
// 	if *configPath != "" {
// 		// запуск с настройкой конфигурации и режима отладки из cli
// 		return createHeartFromConfigFile(*debug)
// 	}
// 	return createHeartFromFlags()
// }

// var createHeartFromConfigFile = func(debug bool) (h types.Heart, err error) {
// 	f, err := os.Open(*configPath)
// 	if err != nil {
// 		return
// 	}

// 	byteValue, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		return
// 	}

// 	c, err := config.NewConfigReader(byteValue, config.ConfigTypeJson)
// 	if err != nil {
// 		return
// 	}

// 	h, err = c.CreateHeart()
// 	if err != nil {
// 		return
// 	}

// 	if isFlagPassed("debug") {
// 		h.Debug = debug // флаг отладки из cli переопределяет отладку в конфигурационном файле
// 	}

// 	return
// }

// var run = func(h types.Heart) {
// 	ctx, cancel := context.WithCancel(context.Background())

// 	engine, err := core.NewEngine(ctx, h)
// 	if err != nil {
// 		exitWithMsg(err.Error())
// 	}

// 	err = engine.Init()
// 	if err != nil {
// 		exitWithMsg(err.Error())
// 	}

// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt)
// 	defer func() {
// 		signal.Stop(c)
// 		cancel()
// 	}()

// 	go func() {
// 		select {
// 		case <-c:
// 			cancel()
// 		case <-ctx.Done():
// 		}
// 	}()

// 	engine.Start()
// }

// var createHeartFromFlags = func() (h types.Heart, err error) {
// 	if *target == "" {
// 		err = fmt.Errorf("Пожалуйста, укажите целевой URL-адрес с помощью флага -t")
// 		return
// 	}

// 	s, err := createScenario()
// 	if err != nil {
// 		return
// 	}

// 	p, err := createProxy()
// 	if err != nil {
// 		return
// 	}

// 	h = types.Heart{
// 		IterationCount:    *iterCount,
// 		LoadType:          strings.ToLower(*loadType),
// 		TestDuration:      *duration,
// 		Scenario:          s,
// 		Proxy:             p,
// 		ReportDestination: *output,
// 		Debug:             *debug,
// 	}
// 	return
// }

// func createProxy() (p proxy.Proxy, err error) {
// 	var proxyURL *url.URL
// 	if *proxyFlag != "" {
// 		proxyURL, err = url.Parse(*proxyFlag)
// 		if err != nil {
// 			return
// 		}
// 	}

// 	p = proxy.Proxy{
// 		Strategy: proxy.ProxyTypeSingle,
// 		Addr:     proxyURL,
// 	}
// 	return
// }

// func createScenario() (s types.Scenario, err error) {
// 	// Auth
// 	var a types.Auth
// 	if *auth != "" {
// 		creds := strings.Split(*auth, ":")
// 		if len(creds) != 2 {
// 			err = fmt.Errorf("не удалось проанализировать учетные данные для проверки подлинности")
// 			return
// 		}

// 		a = types.Auth{
// 			Type:     types.AuthHttpBasic,
// 			Username: creds[0],
// 			Password: creds[1],
// 		}
// 	}

// 	err = types.IsTargetValid(*target)
// 	if err != nil {
// 		return
// 	}

// 	h, err := parseHeaders(headers)
// 	if err != nil {
// 		return
// 	}

// 	step := types.ScenarioStep{
// 		ID:      1,
// 		Method:  strings.ToUpper(*method),
// 		Auth:    a,
// 		Headers: h,
// 		Payload: *payload,
// 		URL:     *target,
// 		Timeout: *timeout,
// 	}

// 	if *certPath != "" && *certKeyPath != "" {
// 		cert, pool, e := types.ParseTLS(*certPath, *certKeyPath)
// 		if e != nil {
// 			err = e
// 			return
// 		}

// 		step.Cert = cert
// 		step.CertPool = pool
// 	}
// 	s = types.Scenario{Steps: []types.ScenarioStep{step}}

// 	return
// }

// func exitWithMsg(msg string) {
// 	if msg != "" {
// 		msg = "err: " + msg
// 		fmt.Fprintln(os.Stderr, msg)
// 	}
// 	os.Exit(1)
// }

// func parseHeaders(headersArr []string) (headersMap map[string]string, err error) {
// 	re := regexp.MustCompile(headerRegexp)
// 	headersMap = make(map[string]string)
// 	for _, h := range headersArr {
// 		matches := re.FindStringSubmatch(h)
// 		if len(matches) < 1 {
// 			err = fmt.Errorf("invalid header:  %v", h)
// 			return
// 		}
// 		headersMap[matches[1]] = matches[2]
// 	}
// 	return
// }

// type header []string

// func (h *header) String() string {
// 	return fmt.Sprintf("%s - %d", *h, len(*h))
// }

// func (h *header) Set(value string) error {
// 	*h = append(*h, value)
// 	return nil
// }

//	func isFlagPassed(name string) bool {
//		found := false
//		flag.Visit(func(f *flag.Flag) {
//			if f.Name == name {
//				found = true
//			}
//		})
//		return found
//	}
package main

import (
	ui "go.ddosify.com/ddosify/frontend"
)

func main() {
	ui.Run()
}
