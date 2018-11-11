package main //LP_L3a buvo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

)



//const CDataFile1 string = "./IFF-6-2_BobinaiteA_L3a_dat_1.txt"
const CDataFile2 string = "./IFF-6-2_BobinaiteA_L3a_dat_2.txt"
//const CDataFile3 string = "./IFF-6-2_BobinaiteA_L3a_dat_3.txt"
const CResultFile string = "./IFF-6-2_BobinaiteA_L3a_rez.txt"

const writersCount int = 5
const readersCount int = 4




const CMaxProcessCount = 5
const CMaxDataCount = 10

var B [CMaxDataCount * CMaxProcessCount]DataStruct
var BCommon Bstruct

type Bstruct struct {
	B             [CMaxDataCount * CMaxProcessCount]DataStruct
	lock          sync.Mutex
	//cond          sync.Cond
	kiekElPridejo int
}

type Data struct { //Automobilio klasė
	name string
	year    int
	price float64
}
type DataStruct struct { //rikiavimo struktūros klasė
	intData int
	count      int
}

type ProcessWriter struct {
	data [CMaxDataCount]Data
	count int
}
type ProcessReader struct {
	data [CMaxDataCount]DataStruct
	count int
}




	func main() {
	fmt.Println("hello world")

		var BCommon Bstruct
		BCommon.kiekElPridejo = 0

		fmt.Println(CDataFile2)

		readFile( CDataFile2)

		printToFileResults(&BCommon, CResultFile)
}

// surenka po duomenu struktura ir issiuncia skaitytojam/rasytojam
func readFile(fileName string){
	var processesWriters [CMaxProcessCount]ProcessWriter
	var processesReaders [CMaxProcessCount]ProcessReader

	var (
		elementNr = 0
		readPElNr = 0
		proccessNr = -1
		readProcessNr = -1
	)

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			values := strings.Split(line, ";")

			if proccessNr == writersCount - 1 && processesWriters[proccessNr].count == elementNr && readProcessNr == -1{ //jei paskutinis writeriu visu elem, tada skaitom readerio kiekį
				kiekis, err := strconv.Atoi(values[0])
				if err != nil {
					fmt.Printf("kiekis=%d, type: %T\n", kiekis,
																kiekis)
				}
				readProcessNr++
				processesReaders[readProcessNr].count = kiekis
			} else {
				if  readProcessNr !=-1 { //jei perskaite visus writersEl tada nebe writers , o readers procesai
					if readProcessNr == readersCount-1 && processesReaders[readProcessNr-1].count == readPElNr{ //jei paskutinis readeris (DS)
						break
					}
					if processesReaders[readProcessNr].count == readPElNr && processesReaders[readProcessNr].count != 0 && readProcessNr < readersCount { //kai read proceso (DS) paskutinis elementas
						readPElNr = 0
						readProcessNr++
						kiekis, err := strconv.Atoi(values[0])
						if  err != nil {
							fmt.Printf("kiekis=%d, type: %T\n", kiekis,
																		kiekis)
						}
						processesReaders[readProcessNr].count = kiekis
					} else {
						if readProcessNr < readersCount && readPElNr < processesReaders[readProcessNr].count {
							intValue, err := strconv.Atoi(values[0])
							if  err != nil {
								fmt.Printf("intValue=%d, type: %T\n", intValue,
									                                         intValue)
							}

							intValue1, err := strconv.Atoi(values[1])
							if  err != nil {
								fmt.Printf("intValue1=%d, type: %T\n", intValue1,
									                                          intValue1)
							}

							processesReaders[readProcessNr].data[readPElNr] = DataStruct{intData: intValue,
																						   count: intValue1}
							readPElNr++
						}
					}
				}
			}

			if proccessNr == -1 && elementNr == 0 { //pirma eilutė writer proceso
				kiekis, err := strconv.Atoi(values[0])
				if err != nil {
					fmt.Printf("kiekis=%d, type: %T\n", kiekis,
																kiekis)
				}
				proccessNr++
				processesWriters[proccessNr].count = kiekis
			}  else {  //kai writer proceso paskutinis elementas
				if processesWriters[proccessNr].count == elementNr && processesWriters[proccessNr].count != 0 && proccessNr < writersCount && readProcessNr == -1 {
					proccessNr++
					elementNr = 0
					kiekis, err := strconv.Atoi(values[0])
					if err != nil {
						fmt.Printf("kiekis=%d, type: %T\n", kiekis,
																	kiekis)
					}
					processesWriters[proccessNr].count = kiekis
				} else { //kol nuskaito visus elementus vieno writre proceso
					if proccessNr < writersCount && elementNr < processesWriters[proccessNr].count && readProcessNr == -1{
						year, err := strconv.Atoi(values[1])
						if err != nil {
							fmt.Printf("year=%d, type: %T\n", year,
								                                     year)
						}

						price1, err := strconv.ParseFloat(values[2], 128)
						if err != nil {
							fmt.Printf("price=%.2f, type: %T\n", price1,
								                                        price1)
						}
						processesWriters[proccessNr].data[elementNr] = Data{name: values[0],
																			 year: year,
																			 price: price1}
						elementNr++
					}
				}
			}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	printToFile(processesWriters, processesReaders, CResultFile)
}


// isprintina pradinius duomenis
func printToFile(P [CMaxProcessCount]ProcessWriter, DS [CMaxProcessCount]ProcessReader, resFileName string) {
	f, err := os.Create(resFileName)
	check(err)
	f.WriteString("--------------------------------------------------------\r\n")
	f.WriteString("============Pradiniai duomenys AUTOMOBILIS============== \r\n")
	f.WriteString("--------------------------------------------------------\r\n")

	i := 0
	lineNo := 0
	for i < writersCount {
		f.WriteString("--------------------------------------------------------\r\n")
		f.WriteString("NR.    PAVADINIMAS             METAI      KAINA    \r\n")
		f.WriteString("--------------------------------------------------------\r\n")
		j := 0
		for j < P[i].count {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\t\t\t" + P[i].data[j].name + "\t\t\t" + strconv.Itoa(P[i].data[j].year) + "\t\t\t" + strconv.FormatFloat(P[i].data[j].price, 'f', 6, 64) + "\r\n")
			j++
		}
		i++
	}

	f.WriteString("-------------------------------------------\r\n")
	f.WriteString("============Pradiniai duomenys DS\r\n==========")
	f.WriteString("-------------------------------------------\r\n")
	i = 0
	lineNo = 0
	for i < readersCount {
		f.WriteString("------------------------------\r\n")
		f.WriteString("NR.    YEAR      INT      \r\n")
		f.WriteString("------------------------------\r\n")
		j := 0
		for j < DS[i].count {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\t" + strconv.Itoa(DS[i].data[j].intData) + "\t" + strconv.Itoa(DS[i].data[j].count) + "\r\n")
			j++
		}
		i++
	}
	f.Close()
}
func printToFileResults(BCommon *Bstruct, resFileName string) {
	B := BCommon.B
	lineNo := 1
	f, err := os.OpenFile(resFileName, os.O_APPEND|os.O_WRONLY, 0600)

	check(err)
	f.WriteString("-------------------------------------------\r\n")
	f.WriteString("REZULTATAI B(bendras masyvas)\r\n")
	f.WriteString("-------------------------------------------\r\n")
	f.WriteString("NR.    STRING          INT         \r\n")
	f.WriteString("--------------------------------------------------------\r\n")

	lineNo = 0
	j := 0
	for j < BCommon.kiekElPridejo {
		if B[j].intData != 0 {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\r\t" +strconv.Itoa(B[j].intData) + "\r\t" + strconv.Itoa(B[j].count) + "\r\n")
		}
		j++
	}
	f.Close()

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}





