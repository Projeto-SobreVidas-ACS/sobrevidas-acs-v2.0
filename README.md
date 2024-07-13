<p align="center">
  <img src="public/assets/images/icone.png" alt="Descrição da imagem" height="200">
</p>

# SobreVidas ACS
### Uma aplicação para o monitoramento do cancer de boca por agentes comunitários da saúde.

## Vídeo com detalhes da implementação e integração com o banco de dados:
link: 
## Vídeo do picth:
link: 

## Instalação:
- Com a linguagem go instalada clone esse repositório;
- Instale PostgreSQL e o pgAdmin 4 e coloque essas configurações:
```
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "5432"=
```
- Dentro do pgAdmin 4 após crie um banco de dados com o nome "sobrevidas-acs";
- Clique com o botão direiro sobre o banco e selecione a ferramentea "query tool"
- Na query execute os seguintes comandos:
```
CREATE TABLE paciente (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100),
    cpf VARCHAR(11) UNIQUE,
    cartao_sus VARCHAR(15) UNIQUE,
    data_nascimento DATE,
    sexo VARCHAR(10),
    unidade_origem VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    celular VARCHAR(15),
    celular2 VARCHAR(15),
    nome_mae VARCHAR(100),
    cep VARCHAR(10),
    cidade VARCHAR(100),
    bairro VARCHAR(100),
    endereco VARCHAR(200),
    tabagista BOOLEAN,
    etilista BOOLEAN,
    lesoes BOOLEAN,
    imagem BYTEA,
    consulta BOOLEAN,
    inativo BOOLEAN DEFAULT false
);

CREATE TABLE usuario (
    id SERIAL PRIMARY KEY,
    senha VARCHAR(100),
    nome VARCHAR(100),
    cpf VARCHAR(11) UNIQUE,
    cbo VARCHAR(10),
    ine VARCHAR(10),
    cnes VARCHAR(10),
    data_nascimento DATE,
    sexo VARCHAR(10),
    email VARCHAR(100) UNIQUE,
    celular VARCHAR(15),
    celular2 VARCHAR(15),
    nome_mae VARCHAR(100),
    cep VARCHAR(10),
    cidade VARCHAR(100),
    bairro VARCHAR(100),
    endereco VARCHAR(200),
    tipo_usuario INTEGER DEFAULT 1,
    inativo BOOLEAN DEFAULT false
);
```
- Banco de dados configurado!
- No terminal, execute o comando go run main.go no repositório;
- Após o log "Conexão bem-sucedida!" acesse a aplicação por http://localhost:8080 no navegador.

## Ferramentas Utilizadas

- **HTML**: Linguagem de marcação para desenvolvimento web.
  - 🌐 [HTML](https://developer.mozilla.org/en-US/docs/Web/HTML)
  
- **CSS**: Linguagem de estilo para design de páginas web.
  - 🎨 [CSS](https://developer.mozilla.org/en-US/docs/Web/CSS)
  
- **GoLang**: Linguagem de programação usada para o desenvolvimento do backend.
  - 👨‍💻 [GoLang](https://golang.org/)
  - Versão utilizada: go 1.22.2
  
- **PostgreSQL**: Sistema de gerenciamento de banco de dados relacional.
  - 🐘 [PostgreSQL](https://www.postgresql.org/)
  - Versão utilizada: PostgreSQL 16.3
  
- **Git**: Sistema de controle de versão distribuído.
  - 📂 [Git](https://git-scm.com/)
  
- **GitHub**: Plataforma de hospedagem de código-fonte e colaboração.
  - 🐙 [GitHub](https://github.com/)
  
- **Figma**: Ferramenta de design de interface de usuário e prototipagem.
  - 🎨 [Figma](https://www.figma.com/)
  
- **Visual Studio Code**: IDE (Ambiente de Desenvolvimento Integrado).
  - 💻 [Visual Studio Code](https://code.visualstudio.com/)

