package main

import (
    "fmt"
)

type countRange struct{
    from int
    to int
}



const top int = 1000000000
//Iterate a variable from x to y
//PFR

func countFromXToY(inp chan countRange, fb chan string){
    for{   
         //Je prends un probl√®me
        rng := <- inp 
         //Je r√souds mon probl√me
        fmt.Printf("IN %d\n", rng.from)
        for cpt := rng.from; cpt < rng.to; cpt++{
            if (cpt%1000 == 0){
                fmt.Printf("%d\n", cpt)
            }

        }
        //PFR    
         //Je pousse mon r√sultat.
        fb <- "FINI"
    }
}

//PFR function pushing jobs
func feedInput(inp chan countRange){
    for mcpt:= 0; mcpt < top; mcpt+= 1000000{
        toPush := countRange{from: mcpt, to: mcpt+999999}
        inp<- toPush
    }
    fmt.Printf("#DEBUG All Pushed\n")
}
func main(){
    //PFR
    var inputChannel chan countRange
    var feedbackChannel chan string

    inputChannel = make(chan countRange, 10)
    feedbackChannel = make(chan string, 10)

    fmt.Printf("#DEBUG START\n")

    for channum := 0; channum < 10; channum++{
        go countFromXToY(inputChannel, feedbackChannel)
    } 

    fmt.Printf("#DEBUG Counting Routines Started\n")
 
    go feedInput(inputChannel) 

    fmt.Printf("#DEBUG PBM Feeding Routine Started\n")

    //countFromXToY(0, 1000000000)

    //PFR
    pushnum := top/1000000
    for rescpt := 0; rescpt < pushnum; rescpt ++{
        fmt.Printf("Pop %d\n", rescpt)
        res := <- feedbackChannel        
         fmt.Printf("Resultat: %v\n", res)
    }
   fmt.Printf("#DEBUG 3_3 END\n")

}
