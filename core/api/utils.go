package api

import (
  "fmt"
  "strings"
)

func ListToString(host string, peers []string) string{
  return fmt.Sprintf("%q,%q", host, strings.Join(peers, ","))
}

func ListFromString(str string) (string, []string){
  splited := strings.Split(str, ",")
  return splited[0], splited[1:]
}
