package src

//"database/sql"
//"fmt"

//"github.com/casbin/casbin/v2"
//casbinpgadapter "github.com/cychiuae/casbin-pg-adapter"

//"casbin-example/enforce"

// var a = mysqladapter.NewAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
// var e = casbin.NewEnforcer("model.conf", a)
//func main() {
//connectionString := "postgresql://postgres:1234@localhost:5432/postgres?sslmode=disable"
//db, err := sql.Open("postgres", connectionString)
//if err != nil {
//	panic(err)
//}

//tableName := "casbin"
//adapter, err := casbinpgadapter.NewAdapter(db, tableName)
//if err != nil {
//	panic(err)
//}

//e, err := casbin.NewEnforcer("model.conf", adapter)
//if err != nil {
//	panic(err)
//}

//requests := map[string][]string{
//	"dang":  {"admin", "/users", "POST"},
//	"phuoc": {"user", "/products", "POST"},
//	"quang": {"user", "/products", "GET"},
//}

//for key, value := range requests {
//fmt.Printf("Check permission for %s\n", key)
//enforce.CasbinEnforce(e, value[0], value[1], value[2])
//}
//}
