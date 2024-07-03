package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "5432"
	dbname   = "sobrevidas-acs"
)

type Paciente struct {
	Nome           string
	CPF            string
	CartaoSUS      string
	DataNascimento string
	Sexo           string
	UnidadeOrigem  string
	Email          string
	Celular1       string
	Celular2       string
	NomeMae        string
	CEP            string
	Cidade         string
	Bairro         string
	Endereco       string
	Tabagista      bool
	Etilista       bool
	Lesoes         bool
	Consulta       bool
}

func main() {
	//BANCO
	// Constrói uma string de conexão
	psqlInfoBanco := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Abre a conexão com o banco de dados
	db, err := sql.Open("postgres", psqlInfoBanco)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verifica a conexão
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	// Print de confirmação para ser retirado
	fmt.Println("Conexão bem-sucedida!")

	// Roteamento para o cadastro
	//http.HandleFunc("/cadastro", cadastroHandler)

	// Inicia o servidor na porta 8080

}
func cadastroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear os dados do formulário
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erro ao parsear formulário", http.StatusBadRequest)
		return
	}

	// Criar um paciente com os dados do formulário
	paciente := Paciente{
		Nome:           r.FormValue("nome"),
		CPF:            r.FormValue("cpf"),
		CartaoSUS:      r.FormValue("cartaoSUS"),
		DataNascimento: r.FormValue("dataNascimento"),
		Sexo:           r.FormValue("sexo"),
		UnidadeOrigem:  r.FormValue("unidadeOrigem"),
		Email:          r.FormValue("email"),
		Celular1:       r.FormValue("telefone1"),
		Celular2:       r.FormValue("telefone2"),
		NomeMae:        r.FormValue("nomeMae"),
		CEP:            r.FormValue("cep"),
		Cidade:         r.FormValue("cidade"),
		Bairro:         r.FormValue("bairro"),
		Endereco:       r.FormValue("enderecoCompleto"),
		Tabagista:      r.FormValue("tabagista") == "true",
		Etilista:       r.FormValue("alcool") == "true",
		Lesoes:         r.FormValue("lesoes") == "true",
		Consulta:       r.FormValue("consulta") == "true",
	}

	// Salvar no banco de dados
	err = salvarNoBanco(paciente)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco de dados", http.StatusInternalServerError)
		return
	}

	// Exemplo de resposta de sucesso
	fmt.Fprintf(w, "Paciente cadastrado com sucesso!")
}

func salvarNoBanco(p Paciente) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Erro ao conectar ao banco de dados: %v", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO paciente (nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem,
                                email, celular, celular2, nome_mae, cep, cidade, bairro, endereco,
                                tabagista, etilista, lesoes, consulta)
                            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`)
	if err != nil {
		log.Printf("Erro ao preparar declaração SQL: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(p.Nome, p.CPF, p.CartaoSUS, p.DataNascimento, p.Sexo, p.UnidadeOrigem,
		p.Email, p.Celular1, p.Celular2, p.NomeMae, p.CEP, p.Cidade, p.Bairro, p.Endereco,
		p.Tabagista, p.Etilista, p.Lesoes, p.Consulta)
	if err != nil {
		log.Printf("Erro ao executar declaração SQL: %v", err)
		return err
	}

	return nil
}
