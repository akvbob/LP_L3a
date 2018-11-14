package main //LP_L3a buvo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)



const CDataFile1 string = "./IFF-6-2_BobinaiteA_L3a_dat_1.txt"
//const CDataFile2 string = "./IFF-6-2_BobinaiteA_L3a_dat_2.txt"
//const CDataFile3 string = "./IFF-6-2_BobinaiteA_L3a_dat_3.txt"
const CResultFile string = "./IFF-6-2_BobinaiteA_L3a_rez.txt"

const writersCount int = 4
const readersCount int = 3
const CMaxProcessCount = 5
const CMaxDataCount = 10


var transferReaderData = make(chan ProcessReader) // perdavimo kanalas tarp failo skaitymo(duomenu perdavimo) ir skaitymo proceso
var transferWriterData = make(chan ProcessWriter) // perdavimo kanalas tarp failo skaitymo(duomenu perdavimo) ir rasymo proceso
var managerW = make(chan ChanStructure)              // kanalas tarp rasymo proceso ir valdytojo proceso
var managerR = make(chan ChanStructure)              // kanalas tarp skaitymo proceso ir valdytojo proceso
var EndChan = make(chan ChanStructure) //pabaigos kanalas
var processesWriters [CMaxProcessCount]ProcessWriter
var processesReaders [CMaxProcessCount]ProcessReader
var wg sync.WaitGroup

type Bstruct struct {
	B             [CMaxDataCount * CMaxProcessCount]DataStruct
	lock          sync.Mutex //kritinės sekcijos apsaugai
	cond          sync.Cond //realizuoti sąlyginę sinchronizaciją - naudojantis COnd galima laukti įvykio ir pranešti apie įvykusį įvykį
	kiekElPridejo int
}

type Car struct { //Automobilio klasė
	name string
	year    int
	price float64
}
type DataStruct struct { //rikiavimo struktūros klasė
	intData int
	count      int
}

type ProcessWriter struct {
	data [CMaxDataCount]Car
	count int
}
type ProcessReader struct {
	data [CMaxDataCount]DataStruct
	count int
}

type ChanStructure struct {
	id int
	data DataStruct
}


func main() {
	fmt.Println("hello world")
	fmt.Println(CDataFile1)
	var BCommon = Bstruct{
		kiekElPridejo: 0,
	}
	readFile(CDataFile1)//, &wg)

	wg.Add(1) //Laukiamų Done kreipinių kiekis - 1
	go duomenuPerdavimoProc(transferWriterData, transferReaderData)
	//wg.Add(1)
	go func() { //valdytojas
		//defer wg.Done()
		for {
			var activeChan  = [] chan ChanStructure{EndChan}
			if BCommon.kiekElPridejo > 0 {
				activeChan = append(activeChan, managerR) //aktyus kanalas - trinimo
			} else {
				activeChan = append(activeChan, nil)
			}
			if BCommon.kiekElPridejo < CMaxDataCount * CMaxProcessCount {
				activeChan = append(activeChan, managerW) //aktyvus kanalas - rašymo
			}else {
				activeChan = append(activeChan, nil)
			}

			var cases [] reflect.SelectCase
			for _, c := range activeChan {
				cases = append(cases, reflect.SelectCase {
					Dir: reflect.SelectRecv,
					Chan: reflect.ValueOf(c),
				})
			}
			choice, _, _ := reflect.Select(cases)
			var message = <- activeChan[choice]
			switch  message.id { //kuris kanalas perduoda žinutę
			case 0: //pabaigos kanalas
				break
			case 1: //rašymo kanalas
				addToB(message, &BCommon)
			case 2: //skaitymo kanalas
				arPasalino := removeFromB(message, &BCommon, managerR)
				//fmt.Printf("ar pasalino" ,message.data, arPasalino)
				fmt.Println(arPasalino)
			}
		}
	}()
	for i := 0; i < CMaxProcessCount; i++ { //rasytojas perduoda valdytojui
		wg.Add(1)
		go rasytojas(transferWriterData, managerW, &wg)
	}
	//wg.Wait() //wait blokuojama, iki bus iškviestas laukiamas kiekis Done kreipinių
	for i := 0; i < CMaxProcessCount; i++ { //skaitytojas perduoda valdytojui duomenis
		wg.Add(1)

		go skaitytojas(transferReaderData, managerR, &wg)
	}

	wg.Wait()
	data1 := DataStruct{
		intData:0,
		count: 0,
	}
	EndChan <- ChanStructure{0,data1}
	printToFileResults(&BCommon, CResultFile)
}

//====================================================Dėjimas į Bendrą masyvą B=========================================
//prideda viena duomenu eilute
func addToB(data ChanStructure, BCommon *Bstruct) {
	BCommon.lock.Lock()
	//fmt.Println(data.data.intData)

	var index int
	index = containsB(data.data.intData, BCommon)
	if index >= 0 {
		BCommon.B[index].count++
	} else {
		index = findAndMakePlace(data.data.intData, BCommon)
		BCommon.B[index] = DataStruct{intData: data.data.intData, count: 1}
	}
	BCommon.kiekElPridejo++

	BCommon.cond.Broadcast() //pažadina visas GO paprogrames , laukiančios naudojantis cond
	BCommon.lock.Unlock()
}

func containsB(value int, BCommon *Bstruct) int {
	j := 0
	for j < len(BCommon.B) {
		if BCommon.B[j].intData == value {
			return j
		}
		j++
	}
	return -1
}

func findAndMakePlace(value int, BCommon *Bstruct) int {
	i := 0
	for BCommon.B[i].intData != 0 && value >= BCommon.B[i].intData {
		i++
	}
	var count int
	count = countOfB(BCommon)
	for count > i {
		BCommon.B[count] = BCommon.B[count-1]
		count--
	}
	return i
}

func countOfB(BCommon *Bstruct) int {
	j := 0
	count := 0
	for j < len(BCommon.B) {
		if BCommon.B[j].intData != 0 {
			count++
		}
		j++
	}
	fmt.Println("Count: ")
	fmt.Println(count)
	return count
}




///==================================================Šalinimas iš bendro masyvo=========================================================
//salina viena duomenu eilute
func removeFromB(data ChanStructure, BCommon *Bstruct, managerR chan ChanStructure) bool { //)
	var arPasalino bool
	arPasalino = false
	BCommon.lock.Lock()
	var index int
	index = containsB(data.data.intData, BCommon)
	if index >= 0 && BCommon.B[index].count > data.data.count {
		BCommon.B[index].count = BCommon.B[index].count - data.data.count
		arPasalino = true
	} else {
		if index >= 0 {
			j := index
			for j < BCommon.kiekElPridejo-1 {
				BCommon.B[j] = BCommon.B[j+1]
				j++
			}
			BCommon.B[j].intData = 0
			BCommon.B[j].count = 0
			BCommon.kiekElPridejo--
			arPasalino = true
		}
	}
	// jei pasalinti nepavyko, bet visi duomenys buvo prideti
	if !arPasalino && BCommon.kiekElPridejo >= CMaxDataCount*CMaxProcessCount {
		arPasalino = true
	}
	// jei pasainti neoavyko, bet i struktura nespeta prideti visu duomenu
	if !arPasalino {
		managerR <- data
		//removeFromB(data, BCommon)
	}
	BCommon.cond.Broadcast()
	BCommon.lock.Unlock()
	return arPasalino
}

//===========================================Procesai===================================================================
// rasytojo proceso metodas
func rasytojas(writeris <-chan ProcessWriter,  managerW chan ChanStructure, wg *sync.WaitGroup) {
	defer wg.Done()
	a := <-writeris
	data := a.data
	for i := 0; i < CMaxDataCount; i++ {
		data1 := DataStruct{ //sukuriu DS iš dviejų laukų tą ir į ją įsidedu reikšmę iš ProcessWriter(rašytojų) duomenų struktūros(Data)
			intData: data[i].year,
			count: 1,
		}
		viens := ChanStructure{id: 1, data: data1 }
		fmt.Println("call add")
		managerW <- viens //perduoda valdytojui duomenis
	}

}
func duomenuPerdavimoProc(transferWriter chan ProcessWriter, transferReader chan ProcessReader) {
	defer wg.Done()
	for i := 0; i < CMaxProcessCount; i++ {
		transferWriter <- processesWriters[i]
	}
	for i := 0; i < CMaxProcessCount; i++ {
		transferReader <- processesReaders[i]
	}
}
// skaitytojo proceso metodas
func skaitytojas(readeris <-chan ProcessReader, managerR chan ChanStructure, wg *sync.WaitGroup) {
	defer wg.Done()
	a := <-readeris
	data := a.data
	for i := 0; i < CMaxDataCount; i++ {
		data1 := DataStruct{
			intData: data[i].intData,
			count: 1,
		}
		viens := ChanStructure{id: 2, data: data1 } ///čia id turi būti 2 , bet su 2 kažkodėl deadlockina....
		fmt.Println("call remove")
		managerR <- viens //perduoda valdytojui duomenis
	}
}

// valdytojo proceso metodas
func valdytojs(managerW chan DataStruct, managerR chan DataStruct, BCommon *Bstruct) {

}
//=============================================================Skaitymas iš failo=======================================
func readFile(fileName string) ([CMaxProcessCount]ProcessWriter,[CMaxProcessCount]ProcessReader){
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
				kiekis, _ := strconv.Atoi(values[0])
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
						kiekis, _ := strconv.Atoi(values[0])
						processesReaders[readProcessNr].count = kiekis
					} else {
						if readProcessNr < readersCount && readPElNr < processesReaders[readProcessNr].count {
							intValue, _ := strconv.Atoi(values[0])
							intValue1, _ := strconv.Atoi(values[1])
							processesReaders[readProcessNr].data[readPElNr] = DataStruct{intData: intValue,
																						   count: intValue1}
							//addToB(processesReaders[readProcessNr].data[readPElNr], BCommon)
							readPElNr++
						}
					}
				}
			}

			if proccessNr == -1 && elementNr == 0 { //pirma eilutė writer proceso
				kiekis, _ := strconv.Atoi(values[0])
				proccessNr++
				processesWriters[proccessNr].count = kiekis
			}  else {  //kai writer proceso paskutinis elementas
				if processesWriters[proccessNr].count == elementNr && processesWriters[proccessNr].count != 0 && proccessNr < writersCount && readProcessNr == -1 {
					proccessNr++
					elementNr = 0
					kiekis, _ := strconv.Atoi(values[0])
					processesWriters[proccessNr].count = kiekis
				} else { //kol nuskaito visus elementus vieno writer proceso
					if proccessNr < writersCount && elementNr < processesWriters[proccessNr].count && readProcessNr == -1{
						year, _ := strconv.Atoi(values[1])
						price1, _ := strconv.ParseFloat(values[2], 128)
						processesWriters[proccessNr].data[elementNr] = Car{name: values[0],
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
	return processesWriters,processesReaders
}

//==============================================Rašymas į failą=========================================================
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
			f.WriteString(strconv.Itoa(lineNo) + "\t\t" +strconv.Itoa(B[j].intData) + "\t\t\t" + strconv.Itoa(B[j].count) + "\r\n")
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





