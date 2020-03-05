package main

import (
	"fmt"
	"net/http"
	"github.com/ajwurts/autodb/developer"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	fmt.Println(r.URL.User)
	fmt.Println("Host", r.URL.Host)
	fmt.Println("Path", r.URL.Path)
	fmt.Println("RawPath", r.URL.RawPath)
	fmt.Println("RawQuery", r.URL.RawQuery)
	fmt.Println("Fragment", r.URL.Fragment)

}


func main() {
	developer.InitializeRemoteDB()
	defer developer.CloseDB()


	user := developer.User{Username:"ajwurts", Password: "password", Token: "token"}
	fmt.Println(user)

	fmt.Println("Create Account: ", developer.CreateAccount("Ajwurts", "password"))

	// Failed Login
	token, err := developer.Login("Ajwurts", "pasword")
	if err != nil {
		fmt.Println(err)
	}


	// Successful Login
	token, err1 := developer.Login("Ajwurts", "password")
	if err != nil {
		fmt.Println(err1)
	}
	fmt.Println("Log In: ", token)

	// Get Developer
	dev, err2 := developer.GetDeveloper(token)
	if err != nil {
		fmt.Println(err2)
	}
	fmt.Println(dev)


	newUser := &developer.User{Username: "Ajwurts", Password: "Password2", Token: token}
	res, err3 := developer.ModifyDeveloper(newUser)
	if err3 != nil {
		fmt.Println(err3)
	}

	fmt.Println(res)

	// Test Modify Developer
	userTest, err4 := developer.GetDeveloperByUsername("Ajwurts")
	if err != nil {
		fmt.Println(err4)
	}
	fmt.Println(userTest)

}