package main
import (
 "fmt"
 "golang.org/x/crypto/bcrypt"
)
func main(){hash := "$2a$10$hI5N7nDwXXRsHGNPNLnbzevJlMnxgL3m50Zude5iMU6F9pYeoCEom"; fmt.Println(bcrypt.CompareHashAndPassword([]byte(hash), []byte("Admin#Pass123")))}
