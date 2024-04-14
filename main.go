package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// struct que transitará os dados do canal
type respAPICEP struct {
	Body *io.ReadCloser
	api  string
}

// main - CLI para buscar CEP
func main() {
	// regex para garantir somente numeros
	var re *regexp.Regexp = regexp.MustCompile("[0-9]+")
	// variavel que vai receber as informações do usuário
	var CEP string

	// interface solicitando CEP
	fmt.Println("Insira um CEP:")
	// populo a variavel CEP com a informação
	_, err := fmt.Scanln(&CEP)
	if err != nil {
		fmt.Println("Ocorreu um erro, não foi possível ler dados de CEP")
		return
	}

	// relizo a pesquisa com regex e faço um join nas informações
	CEP = strings.Join(re.FindAllString(CEP, -1), "")

	//verifico se o tipo está compativel - 00000000
	if len(CEP) != 8 {
		fmt.Println("Ocorreu um erro, CEP inválido")
		return
	}

	// crio um client http e seto o timeout pra evitar fadiga rsrs
	var client http.Client
	client.Timeout = time.Second

	// crio canais pra ambas as pesquisas
	ch := make(chan *respAPICEP)

	// realizo pesquisas cada um em sua rotina
	go searchByCEP("https://brasilapi.com.br/api/cep/v1/"+CEP, &client, ch)
	go searchByCEP("http://viacep.com.br/ws/"+CEP+"/json/", &client, ch)

	// uso select para aguardar as respostas
	// case na primeira que responder ou
	// timeout de 1 segundo
	select {
	case msg := <-ch:
		fmt.Println(fmt.Sprintf("%s:", msg.api))
		fmt.Println(readBody(*msg.Body))
		return
	case <-time.After(time.Second):
		fmt.Println("Consulta excedeu o tempo limite")
		return
	}
}

// searchByCEP - realiza uma busca GET nas URL da documentação
func searchByCEP(URL string, client *http.Client, channel chan *respAPICEP) {
	// realizo GET
	resp, err := client.Get(URL)
	if err != nil {
		fmt.Println("Ocorreu um erro, ao buscar informações")
		close(channel)
		return
	}

	// populo o canal
	channel <- &respAPICEP{Body: &resp.Body, api: URL}

	close(channel)
}

// readBody - realiza a leitura da stream de dados e retorna em string
func readBody(body io.Reader) string {
	b, err := io.ReadAll(body)
	if err != nil {
		fmt.Println("Ocorreu um erro, ao ler informações")
		return ""
	}

	return string(b)
}
