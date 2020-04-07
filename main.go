package main

import (

)

func main(){
  _, err := ParseFlag()
  if err != nil {
    panic(err)
  }
  
}
