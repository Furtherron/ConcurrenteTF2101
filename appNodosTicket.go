package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//variables globales
var bitacora []string //Ips de los nodos de la red
const (
	puerto_registro  = 8000
	puerto_notifica  = 8001
	puerto_procesoHP = 8002
	puerto_solicitud = 8003 //escucha y envío
)

type Vacunar struct {
	GRUPO_RIESGO float64
	EDAD         float64
	SEXO         float64
	DOSIS        float64
	UBIGEO       float64
	FABRICANTE   string
}

var direccionIP_Nodo string

///////////////////////////////////////////////
//estructura de mensaje
type Info struct {
	Tipo     string //tipo de mensaje
	NodeNum  int
	NodeAddr string
}

type MyInfo struct {
	contMsg int    //contar si todos los nodos ya notificaron
	first   bool   //le toca el turno
	nextNum int    //el ticket próximo
	nextApp string //cual es la Ip del ticket próximo
}

var puedeIniciar chan bool //dar el pase para que ejecute la SC
var chMyInfo chan MyInfo
var ticket int

/////////////////////////////////////////////////
//funciones
func ManejadorNotificacion(conn net.Conn) {
	defer conn.Close()
	//leer la notificación
	bufferIn := bufio.NewReader(conn)
	IpNuevoNodo, _ := bufferIn.ReadString('\n')
	IpNuevoNodo = strings.TrimSpace(IpNuevoNodo)
	//actualizar su bitácora
	bitacora = append(bitacora, IpNuevoNodo)
	fmt.Println(bitacora)
}
func AtenderNotificaciones() {
	//modo escucha
	hostlocal := fmt.Sprintf("%s:%d", direccionIP_Nodo, puerto_notifica)
	ln, _ := net.Listen("tcp", hostlocal)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go ManejadorNotificacion(conn)
	}
}

func RegistrarSolicitud(ipConectar string) {
	hostremoto := fmt.Sprintf("%s:%d", ipConectar, puerto_registro)
	conn, _ := net.Dial("tcp", hostremoto)
	defer conn.Close()
	//enviar la Ip del cliente al host remoto
	fmt.Fprintf(conn, "%s\n", direccionIP_Nodo)
	//leer la bitacora que envia el host remoto
	bufferIn := bufio.NewReader(conn)
	msgBitacora, _ := bufferIn.ReadString('\n')
	var arrAuxiliar []string
	json.Unmarshal([]byte(msgBitacora), &arrAuxiliar)
	bitacora = append(arrAuxiliar, ipConectar) //agregar la ip del host remoto a la bitacora del cliente
	fmt.Println(bitacora)
}

func Notificar(ipremoto, ipNuevoNodo string) {
	hostremoto := fmt.Sprintf("%s:%d", ipremoto, puerto_notifica)
	conn, _ := net.Dial("tcp", hostremoto)
	defer conn.Close()
	//enviar la IP del nodo que se este uniendo a la red
	fmt.Fprintf(conn, "%s\n", ipNuevoNodo)
}

func NotificarTodos(ipNuevoNodo string) {
	//recorrer la bitácora y notificar
	for _, dirIp := range bitacora {
		Notificar(dirIp, ipNuevoNodo)
	}
}

func ManejadorSolicitudes(conn net.Conn) {
	defer conn.Close()
	//leer el IP que envia el nodo a unirse a la red
	bufferIn := bufio.NewReader(conn)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip)
	//devolvermos al nodo nuevo la bitacora del nodo actual
	//codificar en formato json la bitacora
	bytesBitacora, _ := json.Marshal(bitacora)
	//serializarlo en string
	fmt.Fprintf(conn, "%s\n", string(bytesBitacora)) //enviar respuesta
	//notificar al resto de nodos de la red del nuevo nodo
	NotificarTodos(ip)
	//actualizar la bitacora del nodo actual
	bitacora = append(bitacora, ip)
	fmt.Println(bitacora) //imprimir la bitácora
}

func AtenderSolicitudRegistro() {
	//modo escucha
	hostlocal := fmt.Sprintf("%s:%d", direccionIP_Nodo, puerto_registro)
	ln, _ := net.Listen("tcp", hostlocal)
	defer ln.Close()
	//atención concurrente
	for {
		conn, _ := ln.Accept() //aceptar las conexiones
		//manejador
		go ManejadorSolicitudes(conn)
	}
}

func EnviarCargaSgteNodo(numero int) {
	//modo envio
	indice := rand.Intn(len(bitacora)) //selecciono de manera aleatoria
	hostremoto := fmt.Sprintf("%s:%d", bitacora[indice], puerto_procesoHP)
	fmt.Printf("Enviando la carga %d al nodo %s\n", numero, bitacora[indice])
	//enviar
	conn, _ := net.Dial("tcp", hostremoto)
	defer conn.Close()
	fmt.Fprintf(conn, "%d\n", numero) //enviar el nro al nodo remoto

}
func ManejadorServicioHP(conn net.Conn) {
	defer conn.Close()
	//leer la carga que llega al nodo
	bufferIn := bufio.NewReader(conn)
	strNum, _ := bufferIn.ReadString('\n')
	strNum = strings.TrimSpace(strNum)
	numero, _ := strconv.Atoi(strNum)
	fmt.Printf("Numero recibido %d\n", numero)
	//lógica del HP
	if numero == 0 {
		fmt.Println("LLegó a su fin, proceso terminado!!!!")
	} else {
		EnviarCargaSgteNodo(numero - 1)
	}

}
func AtenderServicioHP() {
	//modo escucha
	hostlocal := fmt.Sprintf("%s:%d", direccionIP_Nodo, puerto_procesoHP)
	ln, _ := net.Listen("tcp", hostlocal)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go ManejadorServicioHP(conn)
	}
}

//Funciones del servicio de Turnos
func EnviarMensajeSolicitud(addr string, msgInfo Info) {
	addr = strings.TrimSpace(addr)
	//formular el remotehost
	remoteHost := fmt.Sprintf("%s:%d", addr, puerto_solicitud)
	//realizar la llamada
	conn, _ := net.Dial("tcp", remoteHost)
	defer conn.Close()
	//codificar mensaje en formato json
	bytesMsgInfo, _ := json.Marshal(msgInfo)
	//lo serializo a string y se evía
	fmt.Fprintln(conn, string(bytesMsgInfo))
}

func AccederSeccionCritica() {
	fmt.Println("Iniciando el trabajo en sección crítica")
	myInfo := <-chMyInfo
	if myInfo.nextApp == "" {
		fmt.Println("Es el proceso unico existente, y finaliza su trabajo!!")
	} else {
		fmt.Println("Finalizando el trabajo en la sección crítica")
		fmt.Printf("El siguiente nodo a procesar es %s con el ticket %d", myInfo.nextApp, myInfo.nextNum)
		//enviar el mensaje al proximo turno
		msgInfoIni := Info{Tipo: "INICIO"}
		EnviarMensajeSolicitud(myInfo.nextApp, msgInfoIni)
	}
}

//implementación de la lógica del servicio
func ManejadorConexionesSolicitudes(conn net.Conn) {
	defer conn.Close()
	//leer lo enviado
	bufferIn := bufio.NewReader(conn)
	msgInfo, _ := bufferIn.ReadString('\n')
	//descodificar
	var info Info
	json.Unmarshal([]byte(msgInfo), &info)

	fmt.Println(info)

	//lógica de los turnos
	switch info.Tipo {
	case "ENVNUM":
		//sincronizar para acceder y actualiza la info del nodo
		myInfo := <-chMyInfo
		if info.NodeNum < ticket {
			myInfo.first = false //descartamos que sea el próximo
		} else if info.NodeNum < myInfo.nextNum {
			//actualiza
			myInfo.nextNum = info.NodeNum  //actualiza el ticket del proximo turno
			myInfo.nextApp = info.NodeAddr //actualiza el ip del proximo turno
		}
		myInfo.contMsg++ //actualizamos en uno, caad vez que llega un mensaje
		//terminamos de sincronizar la actualización de la info del nodo
		go func() {
			chMyInfo <- myInfo
		}()

		//Evaluar si afirma permiso o no
		//compara si el nodo ya recibió todos los mensajes
		if myInfo.contMsg == len(bitacora) {
			//evalua si es el próximo turno
			if myInfo.first {
				//accede inmediatamente a la sección crítica
				AccederSeccionCritica()
			} else {
				puedeIniciar <- true
			}
		}
	case "INICIO":
		<-puedeIniciar
		AccederSeccionCritica()
	}
}

func AtenderMensajeSolicitud() {
	//modo escucha
	//formular el localHost= IP:PUERTO
	localHost := fmt.Sprintf("%s:%d", direccionIP_Nodo, puerto_solicitud)
	ln, _ := net.Listen("tcp", localHost)
	defer ln.Close()
	//	atención concurrente
	for {
		conn, _ := ln.Accept()
		go ManejadorConexionesSolicitudes(conn)
	}
}

func nodo() {
	direccionIP_Nodo = localAddress()
	fmt.Println("IP: ", direccionIP_Nodo)
	//rol de servidor
	go AtenderSolicitudRegistro()
	go AtenderServicioHP()
	//rol de cliente

	//enviar la solicitud de registro
	bufferIn := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la ip remota: ")
	ipConectar, _ := bufferIn.ReadString('\n')
	ipConectar = strings.TrimSpace(ipConectar)
	//siempre y cuando no sea el primer nodo de la red
	if ipConectar != "" {
		RegistrarSolicitud(ipConectar)
	}

	//rol de servidor
	go AtenderNotificaciones()

	//Servicio de atención de turnos
	//Generar el ticket del nodo
	rand.Seed(time.Now().UTC().UnixNano())
	ticket = rand.Intn(1000000)
	fmt.Println("Ticket=", ticket)

	//Crear los canales
	puedeIniciar = make(chan bool)
	chMyInfo = make(chan MyInfo)

	//Inicializar la info del nodo
	go func() {
		chMyInfo <- MyInfo{0, true, 1000001, ""}
	}()

	//Confirmación de envió de solicitud de turno
	go func() {
		fmt.Print("Presione enter para confirmar el envio de su solicitud de turno...")
		bufferIn := bufio.NewReader(os.Stdin)
		bufferIn.ReadString('\n')
		//Crear el mensaje en base a la estructura Info
		msgInfo := Info{"ENVNUM", ticket, direccionIP_Nodo}
		//notificar al resto de los nodos de la red
		//recorrer la bitácora
		for _, addr := range bitacora {
			//enviar a cada Ip de la bitácora, el mensaje
			go EnviarMensajeSolicitud(addr, msgInfo)
		}
	}()

	//Modo listen de los mensajes de solicitud de turno
	AtenderMensajeSolicitud()
}

func localAddress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("localAddress: %v\n", err.Error()))
		return "127.0.0.1"
	}
	for _, oiface := range ifaces {

		//for _, dir := range oiface.Addrs() {
		//	fmt.Printf("%v %v\n", oiface.Name, dir)
		//}
		//fmt.Println(oiface.Name)

		if strings.HasPrefix(oiface.Name, "Ethernet") {
			addrs, err := oiface.Addrs()
			if err != nil {
				log.Print(fmt.Errorf("localAddress: %v\n", err.Error()))
				continue
			}
			for _, dir := range addrs {
				//fmt.Printf("%v %v\n", oiface.Name, dir)
				switch d := dir.(type) {
				case *net.IPNet:
					//fmt.Println(d.IP)
					if strings.HasPrefix(d.IP.String(), "192") {
						//fmt.Println(d.IP)
						return d.IP.String()
					}

				}
			}
		}
	}
	return "127.0.0.1"
}
