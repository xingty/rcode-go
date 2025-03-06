package main

func main() {
	// fmt.Println("rcode")

	// session := models.SessionPayload[models.SessionParams]{
	// 	Method: "new_session",
	// 	Params: models.SessionParams{
	// 		Pid:      int32(os.Getpid()),
	// 		Hostname: "hostname",
	// 		Keyfile:  "be273bfc-d729-4ded-85f7-ac403e17d71c",
	// 	},
	// }

	// session := models.SessionPayload[models.OpenIDEParams]{
	// 	Method: "open_ide",
	// 	Params: models.OpenIDEParams{
	// 		Sid:  "e8a4e7bb-41a3-48b8-a1ec-2a4d3d0e2b31",
	// 		Path: "/home/darbula/workspace",
	// 		Bin:  "code",
	// 	},
	// }

	// socks := ipc.NewIPCClientSocket("127.0.0.1", 7532)
	// socks.Connect("tcp")

	// jsondata, err := json.Marshal(session)
	// if err != nil {
	// 	panic(err)
	// }

	// err = socks.Send(jsondata)
	// if err != nil {
	// 	panic(err)
	// }

	// response, err := socks.Receive()
	// if err != nil {
	// 	panic(err)
	// }

	// res := models.ResponsePayload[models.SessionData]{}
	// json.Unmarshal(response, &res)
	// fmt.Printf("%s\n", string(response))

	// time.Sleep(100 * time.Minute)
}
