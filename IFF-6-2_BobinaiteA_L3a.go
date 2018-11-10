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
	stringData string
	intData    int
	doubleData float64
}
type DataStruct struct { //rikiavimo struktūros klasė
	stringData string
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
		transferReaderData := make(chan ProcessReader) // perdavimo kanalas tarp failo skaitymo(duomenu perdavimo) ir skaitymo proceso
		transferWriterData := make(chan ProcessWriter) // perdavimo kanalas tarp failo skaitymo(duomenu perdavimo) ir rasymo proceso
		//managerW := make(chan DataStruct)              // kanalas tarp rasymo proceso ir valdytojo proceso
		//managerR := make(chan DataStruct)              // kanalas tarp skaitymo proceso ir valdytojo proceso

		//var wg sync.WaitGroup
		//wg.Add(2)
		go readFile( CDataFile2, transferWriterData, transferReaderData)//&wg,

		/*for i := 0; i < CMaxProcessCount; i++ {
			//wg.Add(1)
			go rasytojas(transferWriterData, managerW, &wg)
		}
		for i := 0; i < CMaxProcessCount; i++ {
			//wg.Add(1)
			go skaitytojas(transferReaderData, managerR, &wg)
		}
		//wg.Add(1)
		go func() {
			defer wg.Done()
			valdytojs(managerW, managerR, &BCommon)
		}()*/

		//wg.Wait()

		printToFileResults(&BCommon, CResultFile)
}

// surenka po duomenu struktura ir issiuncia skaitytojam/rasytojam
func readFile(fileName string, transferWriter chan ProcessWriter, transferReader chan ProcessReader){//,wg *sync.WaitGroup  /*([CMaxProcessCount]ProcessWriter, [CMaxProcessCount]ProcessReader)*/ {

	var processesWriters [CMaxProcessCount]ProcessWriter
	var processesReaders [CMaxProcessCount]ProcessReader



	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, ";")
		for i := 0; i <= writersCount; i++ {
			kiekis, err := strconv.Atoi(values[0])
			if  err != nil {
				fmt.Printf("kiekis=%d, type: %T\n", kiekis, kiekis)
			}

			processesWriters[i].count = kiekis
			for j := 0; j < kiekis ; j++  {
				year, err := strconv.Atoi(values[0])
				if  err != nil {
					fmt.Printf("year=%d, type: %T\n", year, year)
				}

				price, err := strconv.ParseFloat(values[1],64)
				if  err != nil {
					fmt.Printf("price=%d, type: %T\n", price, price)
				}
				var name = values[2]//name
				processesWriters[i].data[j] = Data{intData: year, stringData: name, doubleData: price}
			}
		}

		for i := 0; i <= readersCount; i++ {
			kiekis, err := strconv.Atoi(values[0])
			if  err != nil {
				fmt.Printf("kiekis=%d, type: %T\n", kiekis, kiekis)
			}

			processesReaders[i].count = kiekis
			for j := 0; j < kiekis ; j++  {
				intValue, err := strconv.Atoi(values[0])

				if  err != nil {
					fmt.Printf("intValue=%d, type: %T\n", intValue, intValue)
				}

				var stringValue = values[2]
				processesReaders[i].data[j] = DataStruct{stringData: stringValue, count: intValue}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	printToFile(processesWriters, processesReaders, CResultFile)


	//perduodam rasytojams
	for i := 0; i < CMaxProcessCount; i++ {
		transferWriter <- processesWriters[i]
	}
	//defer wg.Done()
	//perduodam skaitytojams
	for i := 0; i < CMaxProcessCount; i++ {
		transferReader <- processesReaders[i]
	}
	//defer wg.Done()
}


// isprintina pradinius duomenis
func printToFile(P [CMaxProcessCount]ProcessWriter, DS [CMaxProcessCount]ProcessReader, resFileName string) {
	f, err := os.Create(resFileName)
	check(err)
	f.WriteString("-------------------------------------------\r\n")
	f.WriteString("============Pradiniai duomenys AUTOMOBILIS============== \r\n")
	f.WriteString("-------------------------------------------\r\n")

	i := 0
	lineNo := 0
	for i < writersCount {
		f.WriteString("NR.    PAVADINIMAS             METAI      KAINA    \r\n")
		f.WriteString("--------------------------------------------------------\r\n")
		j := 0
		for j < P[i].count {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\r\t" + P[i].data[j].stringData + "\r\t" + strconv.Itoa(P[i].data[j].intData) + "\r\t" + strconv.FormatFloat(P[i].data[j].doubleData, 'f', 6, 64) + "\r\n")
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
		f.WriteString("NR.    STRING              INT.      \r\n")
		f.WriteString("-------------------------------------------\r\n")
		j := 0
		for j < P[i].count {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\r\t" + DS[i].data[j].stringData + "\r\t" + strconv.Itoa(DS[i].data[j].count) + "\r\n")
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
		if B[j].stringData != "" {
			lineNo++
			f.WriteString(strconv.Itoa(lineNo) + "\r\t" + B[j].stringData + "\r\t" + strconv.Itoa(B[j].count) + "\r\n")
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





