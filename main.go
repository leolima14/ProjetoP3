package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"database/sql"
    	"log"
	"os"
	"context"
	"github.com/lib/pq"
)

type Page struct{
	Email string
	Nome string
}

type SeguidorPlaylist struct{
	EmailOuvinte string
	EmailCriador string
	Nome string
}

type Musica struct{
	Id int
	Nome string
	Duracao int
	Curtidas []string
}

type Ouvinte struct{
	Email string
	Password string
	Data_nascimento string
	Nome string
	Sobrenome string
	Telefone string
	Seguindo []string
}

type Playlist struct{
	Email string
	Nome string
	Id_musica []int
	Count int
}

type Artista struct{
	Email string
	Password string
	Data_nascimento string
	Nome_artistico string
	Biografia string
	Ano_formacao int
	Seguidores []string
}

type Data struct {
	Artistas []Artista
	Musicas []Musica
}

type OuvinteMusica struct {
	Ouvintes []Ouvinte
	Musicas []Musica
}

type DataSeguir struct {
	Artistas []Artista
	Ouvintes []Ouvinte
}

type DataSeguidores struct {
	Ar Artista
	Seguidores []string
}

type DataPlaylist struct{
	Playlists []Playlist
	Musicas []Musica
	Ouvintes []Ouvinte
	Artistas []Artista
	Seguidores []SeguidorPlaylist
}

var (
	ctx context.Context
	db  *sql.DB
)

func User(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles("user.html")
		t.Execute(w, nil)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("password")
		var usuario string
		var senha string
		err := db.QueryRow("SELECT email, senha FROM mydb.usuario WHERE email = $1", email).Scan(&usuario, &senha)
		if err != nil {
			fmt.Println("Erro user")
			t, _ := template.ParseFiles("erro.html")
			t.Execute(w, nil)
		}
		if email == usuario && pass == senha{
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else{
			t, _ := template.ParseFiles("user.html")
			t.Execute(w, nil)
		}		
	}
}

func ListaOuvinte(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, _ := db.Query("SELECT * FROM mydb.usuario JOIN mydb.ouvinte USING(email)")
		var ouvintes []Ouvinte
		var artista []Artista
		var email_artista string
		for rows.Next(){
			var email string
			var senha string
			var nome string
			var segue []string
			var data_nascimento string
			var sobrenome string
			var telefone string
			rows.Scan(&email, &senha, &data_nascimento, &nome, &sobrenome, &telefone)
			rows2, _ := db.Query("SELECT email_artista FROM mydb.segue WHERE email_ouvinte = $1", email)
				for rows2.Next(){
					rows2.Scan(&email_artista)
					segue = append(segue, email_artista)
				}
			ouv := Ouvinte{Email: email, Password: senha, Data_nascimento: data_nascimento, Nome: nome, Sobrenome: sobrenome, Telefone: telefone, Seguindo: segue}
			ouvintes = append(ouvintes, ouv)
		}
		rows3, _ := db.Query("SELECT email FROM mydb.artista")
		for rows3.Next(){
			rows3.Scan(&email_artista)
			art := Artista{Email: email_artista}
			artista = append(artista, art)
		}
		data := DataSeguir{Artistas: artista, Ouvintes: ouvintes}
		t, _ := template.ParseFiles("ouvintes.html")
		t.Execute(w, data)
	case "POST":
		valor := r.FormValue("adiciona")
		if valor == "Follow"{
			email_artista := r.FormValue("artistas")
			email_ouvinte := r.FormValue("ouvintes")
			stmt, _ := db.Prepare("INSERT INTO mydb.segue (email_artista, email_ouvinte) values($1, $2)")
			defer stmt.Close()
			stmt.Exec(email_artista, email_ouvinte)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		}else if valor == "Unfollow"{
			email_artista := r.FormValue("artistas")
			email_ouvinte := r.FormValue("ouvintes")
			stmt, _ := db.Prepare("DELETE FROM mydb.segue WHERE email_artista = $1 and email_ouvinte = $2")
			defer stmt.Close()
			stmt.Exec(email_artista, email_ouvinte)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		}
	}
}

func ListaArtista(w http.ResponseWriter, r *http.Request) {
		rows, _ := db.Query("SELECT* FROM mydb.usuario JOIN mydb.artista USING(email)")
		var email string
		var email_seguidor string
		var senha string
		var data_nascimento string
		var artistas []Artista
		for rows.Next(){
			var segue []string
			var nome_artistico string
			var biografia string
			var ano_formacao int
			rows.Scan(&email, &senha, &data_nascimento, &nome_artistico, &biografia, &ano_formacao)
			rows2, _ := db.Query("SELECT email_ouvinte FROM mydb.segue WHERE email_artista = $1", email)
			for rows2.Next(){
				rows2.Scan(&email_seguidor)
				segue = append(segue, email_seguidor)
			}
			art := Artista{Email: email, Password: senha, Data_nascimento: data_nascimento, Nome_artistico: nome_artistico, Biografia: biografia, Ano_formacao: ano_formacao, Seguidores: segue}
			artistas = append(artistas, art)
		}
		t, _ := template.ParseFiles("artistas.html")
		t.Execute(w, artistas)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles("createUser.html")
		t.Execute(w, nil)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("password")
		data_nascimento := r.FormValue("data_nascimento")
		stmt, _ := db.Prepare("INSERT INTO mydb.usuario (email, senha, data_de_nascimento) values($1, $2, $3)")
		defer stmt.Close()
		_, er := stmt.Exec(email, pass, data_nascimento)
		if er != nil {
			t, _ := template.ParseFiles("erro.html")
			t.Execute(w, nil)
		}
		usuario := r.FormValue("tipoUsuario")
		if usuario != "True"{
			telefone := []string{r.FormValue("telefone")}
			nome := r.FormValue("primeiro_nome")
			sobrenome := r.FormValue("sobrenome")
			stmt2, _ := db.Prepare("INSERT INTO mydb.ouvinte (email, primeiro_nome, sobrenome, telefone) values($1, $2, $3, $4)")
			defer stmt2.Close()
			_, er = stmt2.Exec(email, nome, sobrenome, pq.Array(telefone))
			if er != nil {
				t, _ := template.ParseFiles("erro.html")
				t.Execute(w, nil)
			}
		} else{
			artista := r.FormValue("nome_artistico")
			biografia := r.FormValue("biografia")
			ano_formacao := r.FormValue("ano_formacao")
			stmt2, _ := db.Prepare("INSERT INTO mydb.artista (email, nome_artistico, biografia, ano_de_formacao) values($1, $2, $3, $4)")
			_, er = stmt2.Exec(email, artista, biografia, ano_formacao)
			if er != nil {
				t, _ := template.ParseFiles("erro.html")
				t.Execute(w, nil)
			}
		}
		t, _ := template.ParseFiles("user.html")
		t.Execute(w, nil)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("edit.html")
	t.Execute(w, nil)
}

func ListaMusicas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id := r.FormValue("id")
		var nome string
		var duracao int
		var id2 int
		var email string
		if len(id) == 0 {
			rows, _ := db.Query("SELECT* FROM mydb.musica")
			var musicasf []Musica
			var ouvintes []Ouvinte
			for rows.Next(){
				rows.Scan(&id2, &nome, &duracao)
				mus := Musica{Id: id2, Nome: nome, Duracao: duracao}
				musicasf = append(musicasf, mus)
			}
			rows2, _ := db.Query("SELECT email FROM mydb.ouvinte")
			for rows2.Next(){
				rows2.Scan(&email)
				ouv := Ouvinte{Email: email}
				ouvintes = append(ouvintes, ouv)
			}
			data := OuvinteMusica{Ouvintes: ouvintes, Musicas: musicasf}
			t, _ := template.ParseFiles("musicas.html")
			t.Execute(w, data)
		}else {
			id,_ := strconv.Atoi(id)
			var curtidas []string
			db.QueryRow("SELECT nome, duracao FROM mydb.musica WHERE id_musica = $1", id).Scan(&nome, &duracao)
			rows, _ := db.Query("SELECT email FROM mydb.curte WHERE id_musica = $1", id)
			for rows.Next(){
				rows.Scan(&email)
				curtidas = append(curtidas, email)
			}
			mus := Musica{Id: id, Nome: nome, Duracao: duracao, Curtidas: curtidas}
			t, _ := template.ParseFiles("musica.html")
			t.Execute(w, mus)
		}
	case "POST":
		valor := r.FormValue("adiciona")
		if valor == "Like"{
			var id int
			nomeMusica := r.FormValue("musicas")
			emailOuvinte := r.FormValue("Ouvintes")
			db.QueryRow("SELECT id_musica FROM mydb.musica WHERE nome = $1", nomeMusica).Scan(&id)
			stmt, _ := db.Prepare("INSERT INTO mydb.curte (email, id_musica) values($1, $2)")
			defer stmt.Close()
			_, er := stmt.Exec(emailOuvinte, id)
				if er != nil {
					t, _ := template.ParseFiles("erro.html")
					t.Execute(w, nil)
				}
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else if valor == "Unlike"{
			var id int
			nomeMusica := r.FormValue("musicas")
			emailOuvinte := r.FormValue("Ouvintes")
			db.QueryRow("SELECT id_musica FROM mydb.musica WHERE nome = $1", nomeMusica).Scan(&id)
			stmt, _ := db.Prepare("DELETE FROM mydb.curte WHERE email = $1 and id_musica = $2")
			defer stmt.Close()
			_, er := stmt.Exec(emailOuvinte, id)
			if er != nil {
				t, _ := template.ParseFiles("erro.html")
				t.Execute(w, nil)
			}
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		}
	}
}

func GetPlaylist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var nomeOrigem string
		var email string
		var playlists []Playlist
		var musicas []Musica
		var ouvintes []Ouvinte
		var criador []Page
		var users []Artista
		var seguidores []SeguidorPlaylist
		var id_musica int
		var count int
		rows, _ := db.Query("SELECT email FROM mydb.ouvinte")
		for rows.Next(){
			rows.Scan(&email)
			ouv := Ouvinte{Email: email}
			ouvintes = append(ouvintes, ouv)
		}
		rows4, _ := db.Query("SELECT email, nome FROM mydb.playlist")
		for rows4.Next(){
			rows4.Scan(&email, &nomeOrigem)
			p := Page{Email: email, Nome: nomeOrigem}
			criador = append(criador, p)
		}
		for _, o := range criador{
			var ids []int
			rows2, _ := db.Query("SELECT id_musica FROM mydb.musica_playlist WHERE email_cria = $1 and nome = $2", o.Email, o.Nome)
			for rows2.Next(){
				rows2.Scan(&id_musica)
				ids = append(ids, id_musica)
			}
			db.QueryRow("SELECT COUNT(email_ouvinte) FROM mydb.ouvinte_salva_playlist WHERE email_cria = $1 and nome = $2 ", o.Email, o.Nome).Scan(&count)
			prov := Playlist{Email: o.Email, Nome: o.Nome, Id_musica: ids, Count: count}
			playlists = append(playlists, prov)
		}
		rows3, _ := db.Query("SELECT id_musica, nome FROM mydb.musica")
		for rows3.Next(){
			rows3.Scan(&id_musica, &nomeOrigem)
			mus := Musica{Id: id_musica, Nome: nomeOrigem}
			musicas = append(musicas, mus)
		}
		rows1, _ := db.Query("SELECT email FROM mydb.usuario")
		for rows1.Next(){
			rows1.Scan(&email)
			prov := Artista{Email: email}
			users = append(users, prov)
		}
		rows5, _ := db.Query("SELECT * FROM mydb.ouvinte_salva_playlist")
		for rows5.Next(){
			var emailCriador string
			rows5.Scan(&nomeOrigem, &email, &emailCriador)
			prov := SeguidorPlaylist{Nome: nomeOrigem, EmailOuvinte: email, EmailCriador: emailCriador}
			seguidores = append(seguidores, prov)
		}
		data := DataPlaylist{Playlists: playlists, Musicas: musicas, Ouvintes: ouvintes, Artistas: users, Seguidores: seguidores}
		t, _ := template.ParseFiles("playlist.html")
		t.Execute(w, data)
	case "POST":
		valor := r.FormValue("adiciona")
		if valor == "Criar"{
			usuario := r.FormValue("user")
			data := r.FormValue("data_nascimento")
			nome := r.FormValue("nomePlaylist")
			stmt, _ := db.Prepare("INSERT INTO mydb.cria (data_de_criacao, email) values($1, $2)")
			defer stmt.Close()
			stmt.Exec(data, usuario)
			stmt2, _ := db.Prepare("INSERT INTO mydb.playlist (nome, email) values($1, $2)")
			defer stmt2.Close()
			stmt2.Exec(nome, usuario)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else if valor == "Adicionar"{
			musica := r.FormValue("musicas2")
			criador := r.FormValue("Criador1")
			nome := r.FormValue("Nome1")
			var id int
			db.QueryRow("SELECT id_musica FROM mydb.musica WHERE nome = $1", musica).Scan(&id)
			stmt, _ := db.Prepare("INSERT INTO mydb.musica_playlist (nome, id_musica, email_cria) values($1, $2, $3)")
			defer stmt.Close()
			stmt.Exec(nome, id, criador)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else if valor == "Remover"{
			musica := r.FormValue("musicas3")
			criador := r.FormValue("Criador2")
			nome := r.FormValue("Nome2")
			stmt, _ := db.Prepare("DELETE FROM mydb.musica_playlist WHERE nome = $1 and id_musica = $2 and email_cria = $3")
			defer stmt.Close()
			stmt.Exec(nome, musica, criador)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else if valor == "Follow"{
			ouvinte := r.FormValue("ouvintes3")
			criador := r.FormValue("Criador3")
			nome := r.FormValue("Nome")
			stmt, _ := db.Prepare("INSERT INTO mydb.ouvinte_salva_playlist (nome, email_ouvinte, email_cria) values($1, $2, $3)")
			defer stmt.Close()
			stmt.Exec(nome, ouvinte, criador)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		} else if valor == "Unfollow"{
			ouvinte := r.FormValue("ouvintes3")
			criador := r.FormValue("Criador3")
			nome := r.FormValue("Nome")
			stmt, _ := db.Prepare("DELETE FROM mydb.ouvinte_salva_playlist WHERE nome = $1 and email_ouvinte = $2 and email_cria = $3")
			defer stmt.Close()
			stmt.Exec(nome, ouvinte, criador)
			t, _ := template.ParseFiles("edit.html")
			t.Execute(w, nil)
		}
	}
}

func Adiciona(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var artista []Artista
		var nomes []Musica
		var nome string
		rows, _ := db.Query("SELECT nome FROM mydb.musica")
		for rows.Next(){
			rows.Scan(&nome)
			mus := Musica{Id: 0, Nome: nome, Duracao: 0}
			nomes = append(nomes, mus)
		}
		rows, _ = db.Query("SELECT nome_artistico FROM mydb.artista")
		for rows.Next(){
			rows.Scan(&nome)
			art := Artista{Email: "", Password: "",	Data_nascimento: "", Nome_artistico: nome, Biografia: "", Ano_formacao: 0}
			artista = append(artista, art)
		}
		data := Data{Artistas: artista, Musicas:nomes}
		t, _ := template.ParseFiles("adiciona.html")
		t.Execute(w, data)
	case "POST":
		valor := r.FormValue("adiciona")
		if valor == "Adicionar musica"{
			var id int
			rows, _ := db.Query("SELECT id_musica FROM mydb.musica ORDER BY id_musica ASC")
			for rows.Next(){
				rows.Scan(&id)
			}
			id = id + 1
			fmt.Println(id)
			nome := r.FormValue("Nome")
			duracao, _ := strconv.Atoi(r.FormValue("Duracao"))
			stmt2, _ := db.Prepare("INSERT INTO mydb.musica (id_musica, nome, duracao) values($1, $2, $3)")
			defer stmt2.Close()
			_, er := stmt2.Exec(id, nome, duracao)
			if er != nil {
				t, _ := template.ParseFiles("erro.html")
				t.Execute(w, "Title")
			}
		}else if valor == "Alterar"{
			nomeAtual := r.FormValue("musicas")
			novoNome := r.FormValue("nome")
			stmt2, _ := db.Prepare("UPDATE mydb.musica SET nome = $1 WHERE nome = $2")
			defer stmt2.Close()
			_, er := stmt2.Exec(novoNome, nomeAtual)
			if er != nil{
				panic(er)
			}
		}else if valor == "Salvar"{
			m := r.FormValue("musicas2")
			a := r.FormValue("artistas")
			var id int
			var email string
			db.QueryRow("SELECT id_musica FROM mydb.musica WHERE nome = $1", m).Scan(&id)
			db.QueryRow("SELECT email FROM mydb.artista WHERE nome_artistico = $1", a).Scan(&email)
			stmt2, _ := db.Prepare("INSERT INTO mydb.grava (id_musica, email) values($1, $2)")
			defer stmt2.Close()
			_, er := stmt2.Exec(id, email)
			if er != nil{
				panic(er)
			}
		}
		t, _ := template.ParseFiles("edit.html")
    	t.Execute(w, nil)
	}
}

func main() {
	connString := `user=hpyfzpygmsohnm 
	password=b7e93b95a5b4ed4897072f02be5efe46d3a2d07dcf810a90e17cceaed261e01d
	host=ec2-52-4-171-132.compute-1.amazonaws.com
	port=5432
	dbname=dd4vnp4rmq3b8m 
	sslmode=disable`
	db, _ = sql.Open("postgres", connString)

	err := db.Ping()
	if err != nil {
		log.Print("Erro db")
		panic(err)
	}
	
	port:= os.Getenv("PORT")
	http.HandleFunc("/", User)
	http.HandleFunc("/signup", CreateUser)
	http.HandleFunc("/home", Handler)
	http.HandleFunc("/home/musicas", ListaMusicas)
	http.HandleFunc("/home/artistas", ListaArtista)
	http.HandleFunc("/home/ouvintes", ListaOuvinte)
	http.HandleFunc("/home/adiciona", Adiciona)
	http.HandleFunc("/home/playlist", GetPlaylist)
    	log.Fatal(http.ListenAndServe(":" + port, nil))
}
