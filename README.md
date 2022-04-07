# listing
Fast storage and reading of growing lists, and also include the variable's serialization and deserialization. You can easily store simple data such as ID list, or serialize complex data into fixed length data format and store it according to []byte. According to the length, you can easily get the data at the corresponding position and deserialize it.

//Create a new List:

data := []byte{ ... }

//Info is a fixed 26 byte long []byte，but it's contents will change with the addition and deletion operation.

info := listing.StackNew( data )

//Add data to list

info = listing.StackAdd( info, newData )

//Get the specified range data of list:

data = listing.StackGet( info, from, to )

//Get all data in list

data = listing.StackGetAll( info )

//Variable serialize and deserialize

/******************************

Currently supported variable types:

int, uint int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool, string

[]int8, []byte, []int16, []uint16, []int32, []uint32, []int64, []uint64, []float32, []float64, []bool, []int, []uint

Note: []byte decoding requires a set length. Int and uint are serialized to 64 bits.

*******************************/

id := 123

isMan := true

name := "amy"

mobile := "13987654321"

data := listing.VarsToBytes( id, isMan, name, mobile )

var idNew int

var isManNew bool

var name, mobile string

listing.Scan( data, &idNew, &isMan, &name, &mobile )

//In order to unify the data to a fixed length, we can first store all variable length data as info information through listing, and then obtain the original data through info. For example:

articleId := 6545

articleTime := time.Now().UnixNano()

articleTitle := "I have a dream"

articleContent := "I have a dream that one day this nation will rise up and live out the true meaning of its creed: \"We hold these truths to be self-evident, that all men are created equal.\" I have a dream that one day on the red hills of Georgia, the sons of former slaves and the sons of former slave owners will be able to sit down together at the table of brotherhood. I have a dream that one day even the state of Mississippi, a state sweltering with the heat of injustice, sweltering with the heat of oppression, will be transformed into an oasis of freedom and justice. I have a dream that my four little children will one day live in a nation where they will not be judged by the color of their skin but by the content of their character. I have a dream today! I have a dream that one day, down in Alabama, with its vicious racists, with its governor having his lips dripping with the words of \"interposition\" and \"nullification\" -- one day right there in Alabama little black boys and black girls will be able to join hands with little white boys and white girls as sisters and brothers. I have a dream today! I have a dream that one day every valley shall be exalted, and every hill and mountain shall be made low, the rough places will be made plain, and the crooked places will be made straight; \"and the glory of the Lord shall be revealed and all flesh shall see it together.\""

infoTitle := listing.StackNew( []byte(articleTitle) )

infoContent := listing.StackNew( []byte(articleContent) )

data = listing.VarsToBytes( articleId, articleTime, infoTitle, infoContent )

var id2, time2 int

iTitle := make( []byte, 26 )

iContent := make( []byte, 26 )

listing.Scan( &id2, &time2, &iTitle, &iContent )

fmt.Println( "title: ", string( listing.StackGetAll( iTitle ) ) )

fmt.Println( "content: ", string( listing.StackGetAll( iContent ) ) )
