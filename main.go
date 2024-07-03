package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

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
	Imagem         []byte
	Consulta       bool
}

type Usuario struct {
	Senha          string
	Nome           string
	CPF            string
	CBO            string
	INE            string
	CNES           string
	DataNascimento string
	Sexo           string
	Email          string
	Celular        string
	Celular2       string
	NomeMae        string
	CEP            string
	Cidade         string
	Bairro         string
	Endereco       string
	TipoUsuario    int
}

// Função para servir páginas HTML
func renderTemplate(w http.ResponseWriter, tmpl string) {
	tmplPath := filepath.Join("templates", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
} // Manipulador para a página principal
func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html")
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

	// Roteamento para arquivos estáticos
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	// Roteamento para a página principal
	http.HandleFunc("/", homeHandler)

	// Roteamento para o cadastro
	http.HandleFunc("/cadastro", cadastroHandler)
	// Roteamento para o cadastro de usuários
	http.HandleFunc("/cadastro-acs", cadastroAcsHandler)

	// Inicia o servidor na porta 8080
	http.ListenAndServe(":8080", nil)

}
func cadastroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear os dados do formulário
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Erro ao parsear formulário", http.StatusBadRequest)
		log.Printf("Erro ao parsear formulário: %v", err)
		return
	}

	// Inicializar variável para armazenar a imagem
	var fileBytes []byte

	// Verificar se foi enviado um arquivo de imagem
	file, _, err := r.FormFile("imagem")
	if err == nil {
		defer file.Close()

		fileBytes, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Erro ao ler conteúdo do arquivo", http.StatusInternalServerError)
			log.Printf("Erro ao ler conteúdo do arquivo: %v", err)
			return
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, "Erro ao ler arquivo", http.StatusBadRequest)
		log.Printf("Erro ao ler arquivo: %v", err)
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
		Imagem:         fileBytes,
		Consulta:       r.FormValue("consulta") == "true",
	}

	// Salvar no banco de dados
	err = salvarNoBanco(paciente)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco de dados", http.StatusInternalServerError)
		log.Printf("Erro ao salvar no banco de dados: %v", err)
		return
	}

	// Exemplo de resposta de sucesso
	http.Redirect(w, r, "/templates/cadastro-sucesso-paciente.html", http.StatusSeeOther)
}

func salvarNoBanco(p Paciente) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, 5432, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Erro ao conectar ao banco de dados: %v", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO paciente (nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem,
                                email, celular, celular2, nome_mae, cep, cidade, bairro, endereco,
                                tabagista, etilista, lesoes, imagem, consulta)
                            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`)
	if err != nil {
		log.Printf("Erro ao preparar declaração SQL: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(p.Nome, p.CPF, p.CartaoSUS, p.DataNascimento, p.Sexo, p.UnidadeOrigem,
		p.Email, p.Celular1, p.Celular2, p.NomeMae, p.CEP, p.Cidade, p.Bairro, p.Endereco,
		p.Tabagista, p.Etilista, p.Lesoes, p.Imagem, p.Consulta)
	if err != nil {
		log.Printf("Erro ao executar declaração SQL: %v", err)
		return err
	}

	return nil
}
func cadastroAcsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse do formulário
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extrair dados do formulário
	usuario := Usuario{
		Senha:          r.FormValue("password"),
		Nome:           r.FormValue("nome"),
		CPF:            r.FormValue("cpf"),
		CBO:            r.FormValue("cbo"),
		INE:            r.FormValue("ine"),
		CNES:           r.FormValue("cnes"),
		DataNascimento: r.FormValue("dataNascimento"),
		Sexo:           r.FormValue("sexo"),
		Email:          r.FormValue("email"),
		Celular:        r.FormValue("telefone1"),
		Celular2:       r.FormValue("telefone2"),
		NomeMae:        r.FormValue("nomeMae"),
		CEP:            r.FormValue("cep"),
		Cidade:         r.FormValue("cidade"),
		Bairro:         r.FormValue("bairro"),
		Endereco:       r.FormValue("enderecoCompleto"),
		TipoUsuario:    1, // Defina o tipo de usuário conforme necessário
	}

	// Inserir no banco de dados
	err = salvarUsuarioNoBanco(usuario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirecionar ou exibir uma mensagem de sucesso
	http.Redirect(w, r, "/templates/cadastro-sucesso-usuario.html", http.StatusSeeOther)
}
func salvarUsuarioNoBanco(u Usuario) error {
	// Construir a string de conexão
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Abrir a conexão com o banco de dados
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Erro ao conectar ao banco de dados: %v", err)
		return err
	}
	defer db.Close()

	// Preparar a instrução SQL para inserção
	stmt, err := db.Prepare(`INSERT INTO usuario (senha, nome, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`)
	if err != nil {
		log.Printf("Erro ao preparar declaração SQL: %v", err)
		return err
	}
	defer stmt.Close()

	// Executar a instrução SQL
	_, err = stmt.Exec(u.Senha, u.Nome, u.CPF, u.CBO, u.INE, u.CNES, u.DataNascimento, u.Sexo,
		u.Email, u.Celular, u.Celular2, u.NomeMae, u.CEP, u.Cidade, u.Bairro,
		u.Endereco, u.TipoUsuario)
	if err != nil {
		log.Printf("Erro ao executar declaração SQL: %v", err)
		return err
	}

	fmt.Println("Novo usuário inserido com sucesso.")
	return nil
}
