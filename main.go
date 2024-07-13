package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/sessions"

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
	ID             string
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
	ID             string
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

var (
	key   = []byte("PhaFjVgy4TioRUpPNWRYmvdYmZufoV3CJv+IZfVzax8=") //chave
	store = sessions.NewCookieStore(key)
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
	renderTemplateWithData(w, tmpl, nil)
}

func renderTemplateWithData(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := filepath.Join("templates", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html")
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			log.Print("Acesso não autorizado")
			http.Redirect(w, r, "/index", http.StatusSeeOther)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func authenticateUser(db *sql.DB, cpf string, password string) (bool, int, int, error) {
	var storedPassword string
	var userType, userID int

	query := "SELECT senha, tipo_usuario, id FROM usuario WHERE cpf = $1"
	err := db.QueryRow(query, cpf).Scan(&storedPassword, &userType, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, 0, nil
		}
		return false, 0, 0, err
	}

	if password != storedPassword {
		return false, 0, 0, nil
	}

	return true, userType, userID, nil
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			renderTemplate(w, "login.html")
			return
		}

		if r.Method == http.MethodPost {
			cpf := r.FormValue("username")
			password := r.FormValue("password")

			authenticated, userType, userID, err := authenticateUser(db, cpf, password)
			if err != nil {
				http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
				return
			}

			if authenticated {
				session, _ := store.Get(r, "session-name")
				session.Values["authenticated"] = true
				session.Values["userType"] = userType
				session.Values["userID"] = userID
				session.Save(r, w)

				switch userType {
				case 1:
					http.Redirect(w, r, "templates/dashboard-acs.html", http.StatusSeeOther)
				case 2:
					http.Redirect(w, r, "templates/dashboard-adm.html", http.StatusSeeOther)
				default:
					http.Error(w, "CPF ou senha incorretos", http.StatusUnauthorized)
				}
			} else {
				http.Error(w, "CPF ou senha incorretos", http.StatusUnauthorized)
			}
		}
	}
}
func backHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Acesso não autorizado", http.StatusUnauthorized)
		return
	}

	userType := session.Values["userType"]
	url := r.URL.Path

	if userType == 1 {
		http.Redirect(w, r, "/templates/dashboard-acs.html", http.StatusSeeOther)
		return
	}
	if userType == 2 {
		http.Redirect(w, r, "/templates/dashboard-adm.html", http.StatusSeeOther)
		return
	}

	renderTemplate(w, url)
}

func cadastroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Erro ao parsear formulário", http.StatusBadRequest)
		log.Printf("Erro ao parsear formulário: %v", err)
		return
	}

	//variável para armazenar a imagem
	var fileBytes []byte

	// verificar se foi enviado imagem
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

	// cria paciente com os dados do formulário
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

	err = salvarNoBanco(paciente)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco de dados", http.StatusInternalServerError)
		log.Printf("Erro ao salvar no banco de dados: %v", err)
		return
	}

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
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tipoUsuarioStr := r.FormValue("tipo")
	tipoUsuario, err := strconv.Atoi(tipoUsuarioStr)
	if err != nil {
		http.Error(w, "Tipo de usuário inválido", http.StatusBadRequest)
		return
	}

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
		TipoUsuario:    tipoUsuario,
	}

	err = salvarUsuarioNoBanco(usuario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/templates/cadastro-sucesso-usuario.html", http.StatusSeeOther)
}
func salvarUsuarioNoBanco(u Usuario) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Erro ao conectar ao banco de dados: %v", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO usuario (senha, nome, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`)
	if err != nil {
		log.Printf("Erro ao preparar declaração SQL: %v", err)
		return err
	}
	defer stmt.Close()

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

func getPacienteByCPF(db *sql.DB, cpf string) ([]Paciente, error) {
	query := `
        SELECT id, nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tabagista, etilista, lesoes, imagem, consulta 
        FROM paciente
        WHERE cpf = $1 AND inativo = false
    `
	var paciente1 []Paciente
	var p Paciente
	err := db.QueryRow(query, cpf).Scan(&p.ID, &p.Nome, &p.CPF, &p.CartaoSUS, &p.DataNascimento, &p.Sexo, &p.UnidadeOrigem, &p.Email, &p.Celular1, &p.Celular2, &p.NomeMae, &p.CEP, &p.Cidade, &p.Bairro, &p.Endereco, &p.Tabagista, &p.Etilista, &p.Lesoes, &p.Imagem, &p.Consulta)
	if err != nil {
		log.Printf("Erro ao fazer scan dos resultados: %v", err)
		return nil, err
	}
	paciente1 = append(paciente1, p)

	return paciente1, nil
}

func getPacienteIDByCPF(db *sql.DB, cpf string) (int, error) {
	var id int
	query := "SELECT id FROM paciente WHERE cpf = $1"
	err := db.QueryRow(query, cpf).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Paciente com CPF %s não encontrado", cpf)
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func identificarEAtualizarHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cpf := r.FormValue("identificar-cpf")
		if cpf == "" {
			http.Error(w, "CPF é obrigatório", http.StatusBadRequest)
			return
		}

		id, err := getPacienteIDByCPF(db, cpf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/editar-paciente?id=%d", id), http.StatusSeeOther)
	}
}

func getPacientes(db *sql.DB) ([]Paciente, error) {
	query := `
        SELECT id, nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tabagista, etilista, lesoes, imagem, consulta 
        FROM paciente
        WHERE inativo = false
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erro ao executar consulta: %v", err)
		return nil, err
	}
	defer rows.Close()

	var pacientes []Paciente
	for rows.Next() {
		var p Paciente
		err := rows.Scan(&p.ID, &p.Nome, &p.CPF, &p.CartaoSUS, &p.DataNascimento, &p.Sexo, &p.UnidadeOrigem, &p.Email, &p.Celular1, &p.Celular2, &p.NomeMae, &p.CEP, &p.Cidade, &p.Bairro, &p.Endereco, &p.Tabagista, &p.Etilista, &p.Lesoes, &p.Imagem, &p.Consulta)
		if err != nil {
			log.Printf("Erro ao fazer scan dos resultados: %v", err)
			return nil, err
		}
		pacientes = append(pacientes, p)
	}

	log.Printf("Número de pacientes encontrados: %d", len(pacientes))
	return pacientes, nil
}

func listarPacientesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cpf := r.FormValue("buscar-cpf")
		if cpf != "" {
			paciente, err := getPacienteByCPF(db, cpf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			renderTemplateWithData(w, "lista-pacientes.html", paciente)
			return
		}

		// listar todos os pacientes
		pacientes, err := getPacientes(db)
		if err != nil {
			log.Printf("Erro ao obter pacientes: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Renderizando template com %d pacientes", len(pacientes))
		renderTemplateWithData(w, "lista-pacientes.html", pacientes)
	}
}

func getUsuarioByCPF(db *sql.DB, cpf string) ([]Usuario, error) {
	query := `
        SELECT id, nome, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario 
		FROM usuario
        WHERE cpf = $1 AND inativo = false
    `
	var usuario1 []Usuario
	var u Usuario
	err := db.QueryRow(query, cpf).Scan(&u.ID, &u.Nome, &u.CPF, &u.CBO, &u.INE, &u.CNES, &u.DataNascimento, &u.Sexo,
		&u.Email, &u.Celular, &u.Celular2, &u.NomeMae, &u.CEP, &u.Cidade, &u.Bairro,
		&u.Endereco, &u.TipoUsuario)

	if err != nil {
		log.Printf("Erro ao fazer scan dos resultados: %v", err)
		return nil, err
	}
	usuario1 = append(usuario1, u)

	return usuario1, nil
}

func getUsuarios(db *sql.DB) ([]Usuario, error) {
	query := `
        SELECT id, nome, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario 
		FROM usuario
        WHERE inativo = false
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erro ao executar consulta: %v", err)
		return nil, err
	}
	defer rows.Close()

	var usuarios []Usuario
	for rows.Next() {
		var u Usuario
		err := rows.Scan(&u.ID, &u.Nome, &u.CPF, &u.CBO, &u.INE, &u.CNES, &u.DataNascimento, &u.Sexo,
			&u.Email, &u.Celular, &u.Celular2, &u.NomeMae, &u.CEP, &u.Cidade, &u.Bairro,
			&u.Endereco, &u.TipoUsuario)
		if err != nil {
			log.Printf("Erro ao fazer scan dos resultados: %v", err)
			return nil, err
		}
		usuarios = append(usuarios, u)
	}

	log.Printf("Número de usuarios encontrados: %d", len(usuarios))
	return usuarios, nil
}

func listarUsuariosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// verifica se tem cpf procurando
		cpf := r.FormValue("buscar-cpf-user")
		if cpf != "" {
			usuario, err := getUsuarioByCPF(db, cpf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			renderTemplateWithData(w, "lista-usuarios.html", usuario)
			return
		}

		usuarios, err := getUsuarios(db)
		if err != nil {
			log.Printf("Erro ao obter usuarios: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Renderizando template com %d usuarios", len(usuarios))
		renderTemplateWithData(w, "lista-usuarios.html", usuarios)
	}
}

func editarPacienteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		var paciente Paciente

		err := db.QueryRow("SELECT id, nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tabagista, etilista, lesoes, imagem, consulta FROM paciente WHERE id = $1", id).Scan(
			&paciente.ID, &paciente.Nome, &paciente.CPF, &paciente.CartaoSUS, &paciente.DataNascimento, &paciente.Sexo, &paciente.UnidadeOrigem, &paciente.Email, &paciente.Celular1, &paciente.Celular2, &paciente.NomeMae, &paciente.CEP, &paciente.Cidade, &paciente.Bairro, &paciente.Endereco, &paciente.Tabagista, &paciente.Etilista, &paciente.Lesoes, &paciente.Imagem, &paciente.Consulta)

		if err != nil {
			http.Error(w, "Paciente não encontrado", http.StatusNotFound)
			return
		}

		parsedTime, err := time.Parse(time.RFC3339, paciente.DataNascimento)
		if err != nil {
			http.Error(w, "Erro ao fazer parse da data", http.StatusInternalServerError)
			return
		}

		DataFormat := parsedTime.Format("2006-01-02")
		paciente.DataNascimento = DataFormat

		tmpl, err := template.ParseFiles("templates/atualizar-paciente.html")
		log.Print(paciente.DataNascimento)
		if err != nil {
			http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, paciente)
	}
}

func atualizarPacienteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Erro ao analisar o formulário: "+err.Error(), http.StatusBadRequest)
			return
		}

		id := r.FormValue("id")
		nome := r.FormValue("nome")

		cartaoSUS := r.FormValue("cartaoSUS")
		dataNascimento := r.FormValue("dataNascimento")
		sexo := r.FormValue("sexo")
		unidadeOrigem := r.FormValue("unidadeOrigem")
		email := r.FormValue("email")
		celular := r.FormValue("telefone1")
		celular2 := r.FormValue("telefone2")
		nomeMae := r.FormValue("nomeMae")
		cep := r.FormValue("cep")
		cidade := r.FormValue("cidade")
		bairro := r.FormValue("bairro")
		endereco := r.FormValue("enderecoCompleto")
		tabagista := r.FormValue("tabagista") == "true"
		etilista := r.FormValue("etilista") == "true"
		lesoes := r.FormValue("lesoes") == "true"
		consulta := r.FormValue("consulta") == "true"

		log.Print(dataNascimento)

		var imagem []byte
		file, _, err := r.FormFile("imagem")
		if err == nil {
			defer file.Close()
			imagem, err = io.ReadAll(file)
			if err != nil {
				http.Error(w, "Erro ao ler a imagem: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		_, err = db.Exec(`
            UPDATE paciente
            SET nome = $1, cartao_sus = $2, data_nascimento = $3, sexo = $4, unidade_origem = $5,
                email = $6, celular = $7, celular2 = $8, nome_mae = $9, cep = $10, cidade = $11,
                bairro = $12, endereco = $13, tabagista = $14, etilista = $15, lesoes = $16, imagem = $17, consulta = $18
            WHERE id = $19`,
			nome, cartaoSUS, dataNascimento, sexo, unidadeOrigem, email, celular, celular2,
			nomeMae, cep, cidade, bairro, endereco, tabagista, etilista, lesoes, imagem, consulta, id,
		)
		if err != nil {
			http.Error(w, "Erro ao atualizar paciente: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/templates/atualizar-sucesso-paciente.html", http.StatusSeeOther)
	}
}

func editarUsuarioHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		var usuario Usuario

		err := db.QueryRow("SELECT id, nome, senha, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario FROM usuario WHERE id = $1", id).Scan(
			&usuario.ID, &usuario.Nome, &usuario.Senha, &usuario.CPF, &usuario.CBO, &usuario.INE, &usuario.CNES, &usuario.DataNascimento, &usuario.Sexo,
			&usuario.Email, &usuario.Celular, &usuario.Celular2, &usuario.NomeMae, &usuario.CEP, &usuario.Cidade, &usuario.Bairro,
			&usuario.Endereco, &usuario.TipoUsuario)

		if err != nil {
			http.Error(w, "Usuario não encontrado", http.StatusNotFound)
			return
		}

		parsedTime, err := time.Parse(time.RFC3339, usuario.DataNascimento)
		if err != nil {
			http.Error(w, "Erro ao fazer parse da data", http.StatusInternalServerError)
			return
		}

		DataFormat := parsedTime.Format("2006-01-02")
		usuario.DataNascimento = DataFormat

		tmpl, err := template.ParseFiles("templates/atualizar-acs.html")
		log.Print(usuario.DataNascimento)
		if err != nil {
			http.Error(w, "Erro ao carregar o template", http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, usuario)
	}
}

func atualizarUsuarioHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := r.FormValue("id")
		nome := r.FormValue("nome")
		senha := r.FormValue("password")

		cbo := r.FormValue("cbo")
		ine := r.FormValue("ine")
		cnes := r.FormValue("cnes")
		dataNascimento := r.FormValue("dataNascimento")
		sexo := r.FormValue("sexo")
		email := r.FormValue("email")
		celular := r.FormValue("telefone1")
		celular2 := r.FormValue("telefone2")
		nomeMae := r.FormValue("nomeMae")
		cep := r.FormValue("cep")
		cidade := r.FormValue("cidade")
		bairro := r.FormValue("bairro")
		endereco := r.FormValue("enderecoCompleto")

		tipoUsuarioStr := r.FormValue("tipo")
		tipoUsuario, err := strconv.Atoi(tipoUsuarioStr)
		if err != nil {
			http.Error(w, "Tipo de usuário inválido", http.StatusBadRequest)
			return
		}

		log.Printf("ID: %s, TipoUsuario: %d", id, tipoUsuario)
		log.Print(dataNascimento)

		_, err = db.Exec(`
            UPDATE usuario
            SET nome = $1, senha = $2, data_nascimento = $3, sexo = $4, tipo_usuario = $5,
                email = $6, celular = $7, celular2 = $8, nome_mae = $9, cep = $10, cidade = $11,
                bairro = $12, endereco = $13, cbo = $14, ine = $15, cnes = $16
            WHERE id = $17`,
			nome, senha, dataNascimento, sexo, tipoUsuario, email, celular, celular2,
			nomeMae, cep, cidade, bairro, endereco, cbo, ine, cnes, id,
		)
		if err != nil {
			http.Error(w, "Erro ao atualizar paciente: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/templates/atualizar-sucesso-usuario.html", http.StatusSeeOther)
	}
}
func getUsuarioByID(db *sql.DB, id int) (Usuario, error) {
	query := `
        SELECT id, nome, cpf, cbo, ine, cnes, data_nascimento, sexo, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tipo_usuario 
		FROM usuario
        WHERE id = $1 AND inativo = false
    `
	var u Usuario
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Nome, &u.CPF, &u.CBO, &u.INE, &u.CNES, &u.DataNascimento, &u.Sexo,
		&u.Email, &u.Celular, &u.Celular2, &u.NomeMae, &u.CEP, &u.Cidade, &u.Bairro,
		&u.Endereco, &u.TipoUsuario)

	if err != nil {
		if err == sql.ErrNoRows {
			return u, nil
		}
		return u, err
	}
	parsedTime, err := time.Parse(time.RFC3339, u.DataNascimento)
	if err != nil {
		return u, fmt.Errorf("erro ao fazer parse da data: %v", err)
	}

	u.DataNascimento = parsedTime.Format("2006-01-02")

	return u, nil
}
func perfilHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Acesso não autorizado", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["userID"].(int)
		if !ok {
			http.Error(w, "ID do usuário não encontrado na sessão", http.StatusInternalServerError)
			return
		}

		usuario, err := getUsuarioByID(db, userID)
		if err != nil {
			http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
			return
		}

		renderTemplateWithData(w, "/atualizar-perfil.html", usuario)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	session.Values["authenticated"] = false
	session.Values["userType"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

// filtros
func getPacientesConsultas(db *sql.DB) ([]Paciente, error) {
	query := `
        SELECT id, nome, cpf, cartao_sus, data_nascimento, sexo, unidade_origem, email, celular, celular2, nome_mae, cep, cidade, bairro, endereco, tabagista, etilista, lesoes, imagem, consulta 
        FROM paciente
        WHERE inativo = false AND consulta = true
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erro ao executar consulta: %v", err)
		return nil, err
	}
	defer rows.Close()

	var pacientes []Paciente
	for rows.Next() {
		var p Paciente
		err := rows.Scan(&p.ID, &p.Nome, &p.CPF, &p.CartaoSUS, &p.DataNascimento, &p.Sexo, &p.UnidadeOrigem, &p.Email, &p.Celular1, &p.Celular2, &p.NomeMae, &p.CEP, &p.Cidade, &p.Bairro, &p.Endereco, &p.Tabagista, &p.Etilista, &p.Lesoes, &p.Imagem, &p.Consulta)
		if err != nil {
			log.Printf("Erro ao fazer scan dos resultados: %v", err)
			return nil, err
		}
		pacientes = append(pacientes, p)
	}

	log.Printf("Número de pacientes encontrados: %d", len(pacientes))
	return pacientes, nil
}

func listarPacientesConsultasHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cpf := r.FormValue("buscar-cpf")
		if cpf != "" {
			paciente, err := getPacienteByCPF(db, cpf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			renderTemplateWithData(w, "lista-consultas.html", paciente)
			return
		}

		// listar todos os pacientes
		pacientes, err := getPacientesConsultas(db)
		if err != nil {
			log.Printf("Erro ao obter pacientes: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Renderizando template com %d consultas", len(pacientes))
		renderTemplateWithData(w, "lista-consultas.html", pacientes)
	}
}

func main() {
	//BANCO
	// string de conexão
	psqlInfoBanco := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	// abre a conexão
	db, err := sql.Open("postgres", psqlInfoBanco)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// verifica a conexão
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	// print de confirmação para ser retirado ou n
	fmt.Println("Conexão bem-sucedida!")

	// para abrir arquivos estáticos
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	// ativa a página principal
	http.HandleFunc("/", homeHandler)
	// ativa a função de login
	http.HandleFunc("/login", loginHandler(db))
	http.HandleFunc("/logout", logoutHandler)

	// ativa o cadastro
	http.HandleFunc("/cadastro", requireAuth(cadastroHandler))
	// identificar e atualizar usuário
	http.HandleFunc("/paciente-existente", requireAuth(identificarEAtualizarHandler(db)))

	// para editar paciente
	http.HandleFunc("/editar-paciente", requireAuth(editarPacienteHandler(db)))
	// atualizar paciente
	http.HandleFunc("/atualizar-paciente", requireAuth(atualizarPacienteHandler(db)))

	// para editar usuario
	http.HandleFunc("/editar-usuario", requireAuth(editarUsuarioHandler(db)))
	// atualizar usuario
	http.HandleFunc("/atualizar-usuario", requireAuth(atualizarUsuarioHandler(db)))
	http.HandleFunc("/perfil-usuario", requireAuth(perfilHandler(db)))

	// ativa o cadastro de usuários
	http.HandleFunc("/cadastro-acs", requireAuth(cadastroAcsHandler))
	//ativa listar paciente
	http.HandleFunc("/lista-pacientes", requireAuth(listarPacientesHandler(db)))
	http.HandleFunc("/lista-usuarios", requireAuth(listarUsuariosHandler(db)))
	//filtros
	http.HandleFunc("/lista-consultas", requireAuth(listarPacientesConsultasHandler(db)))

	//direciona para voltar
	http.HandleFunc("/voltar", requireAuth(backHandler))

	// inicia o servidor na porta 8080
	http.ListenAndServe(":8080", nil)

}
