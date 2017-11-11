package main
import (
    "encoding/json"
    "net/http"
    "core"
    "github.com/julienschmidt/httprouter"
)

func main() {
    router: = httprouter.New()
    router.GET("/getuser", getUser)
    log.Fatal(http.ListenAndServe(":8080", router))
    var GetAll = core.Decorate(
        core.HandlerFunc(getUser),
        core.Cache(reflect.TypeOf([] * User {}), cache.UserExpiration),
        core.Search([] string {
            "Email", "Name"
        }),
        core.Paginate("Company"),
        core.PrivateKey(util.PrivateKey())
    )
}

func getUser(r * http.Request, ps httprouter.Params, username string)(interface {}, error) {

    resp, err: = http.Get("http://www.json-generator.com/api/json/get/cpIuGuaIya?indent=2");

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var User struct {
        Index int64 `json:"index"`
        Id int64 `json:"_id"`
        Name string `json:"name"`
        Gender string `json:"gender"`
        Age int64 `json:"age"`
        Email string `json:"email"`
        EyeColor string `json:"eyeColor"`
        Phone string `json:"phone"`
        Address string `json:"address"`
        Balance string `json:"balance"`
        Guid string `json:"guid"`
        Company string `json:"company"`
        isActive bool `json:"isActive"`
    }

    if err: = json.NewDecoder(resp.Body).Decode( & User);
    err != nil {
        return nil, err
    }

    return User, nil
}