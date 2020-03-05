package developer

import "fmt"

type User struct  {
	Username string
	Password string
	Token string
	loggedIn bool
}


type LoginError string 

func (e LoginError) Error() string {
	return fmt.Sprintf("failed to log in: %s", string(e))
}

type UserNotFoundError string

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("User Not Found: %s", string(e))
}

var data map[string]*User = make(map[string]*User)
var tokenKey map[string]*User = make(map[string]*User)

func generateToken() string {
	return "12345678910"
}

func Login(Username, Password string) (string, error) {
	if usr, ok := data[Username]; ok {
		if usr.Password == Password && !usr.loggedIn {
			usr.loggedIn = true
			token := generateToken()
			usr.Token = token
			tokenKey[token] = usr
			return token, nil

		} else {
			return "", LoginError("Bad Username or Password")
		}
	}
	return "", LoginError("Bad Username or Password")
}

func Logout(Username string) bool {
	if usr, ok := data[Username]; ok {
		usr.loggedIn = false
		return true
	}
	return false
}

func CreateAccount(Username, Password string) bool {
	if _, ok := data[Username]; ok {
		return false
	} else {
		user := &User{Username: Username, Password: Password}
		data[Username] = user
		return true
	}

}

func GetDeveloper(token string) (*User, error) {
	if usr, ok := tokenKey[token]; ok {
		return usr, nil
	} else {
		return &User{}, UserNotFoundError("No User Connected to Token")
	}
}

func GetDeveloperByUsername(username string) (*User, error) {
	if usr, ok := data[username]; ok {
		return usr, nil
	} else {
		return &User{}, UserNotFoundError("No User Connected to Token")
	}
}


func ModifyDeveloper(newUser *User) (bool, error) {
	if usr, ok := tokenKey[newUser.Token]; ok {
		*usr = *newUser
		return true, nil
	}
	return false, UserNotFoundError(newUser.Username)
}

func DeleteDeveloper(token string) bool {
	if usr, ok := tokenKey[token]; ok {
		username := usr.Username
		delete(tokenKey, token)
		delete(data, username)
		return true
	} else {
		return false
	}
}

func (u User) String() string {
	return fmt.Sprintf("%s, %s, %s", u.Username, u.Password, u.Token)
}
