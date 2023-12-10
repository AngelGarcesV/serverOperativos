package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func clearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Println("Sistema operativo no compatible con limpieza de pantalla.")
	}
}

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		fmt.Println("Se esperan al menos tres argumentos.")
		return
	}

	direccionIP := args[0]
	Puerto := args[1]
	direccionPuerto := direccionIP + ":" + Puerto

	var socket net.Conn
	fmt.Println("=========================")
	fmt.Println("Cliente Operativos 2023.2")
	fmt.Println("=========================")
	var bye bool = false

	tcpAddressC, _ := net.ResolveTCPAddr("tcp4", direccionPuerto)
	var err error
	socket, err = net.DialTCP("tcp", nil, tcpAddressC)
	if err != nil {
		clearScreen()
		fmt.Println("Error al establecer la conexión")
		return
	} else {
		fmt.Println("client# Se ha establecido la conexion con(", socket.RemoteAddr(), ")")
	}

	enviarTiempo(&socket, args[2])

	for {
		fmt.Println("Elija que desea hacer:\n1)Registrarse\n2)Iniciar Sesion\n3)Salir")
		lector := bufio.NewReader(os.Stdin)
		i, err := lector.ReadString('\n')
		i = strings.TrimRight(i, "\r\n")
		if err != nil {
			clearScreen()
			fmt.Println("Eleccion erronea")
		}
		if i != "1" && i != "2" && i != "3" {
			clearScreen()
			continue
		}
		if i == "3" {
			break
		}
		if i == "1" || i == "2" {
			enviarOpcion(&socket, i)
			clearScreen()
			for {
				if i == "1" {
					var mssg string = ""
					for {
						fmt.Println("Por favor ingrese el nombre de usuario a registrar:")
						name, err := lector.ReadString('\n')
						if err != nil {
							clearScreen()
							fmt.Println("Nombre inadecuado")
						}
						usuario := name
						usuario = strings.TrimRight(usuario, "\r\n")
						fmt.Println("Por favor ingrese la contraseña para su cuenta:")
						pss1 := PedirPassword()

						fmt.Println("Por favor ingrese la contraseña nuevamente:")
						pss2 := PedirPassword()

						if pss1 == pss2 {
							mssg = usuario + "|" + pss1
							fmt.Println("Esperando aceptación de registro por parte del servidor...")
							enviarRegistro(&socket, &mssg)
							if mssg == "fail" || mssg == "fin" {
								fmt.Println("El registro ha fallado:")
								break
							} else {
								clearScreen()
								fmt.Println("Registro exitoso:")
								break
							}
						} else {
							clearScreen()
							fmt.Println("Las contraseñas no coinciden")
							mssg = "fail"
							break
						}
					}
					if mssg == "fail" {
						continue
					}
					if mssg == "fin" {
						bye = true
						break
					}
					break
				} else {
					reader := bufio.NewReader(socket)
					intentos, _ := reader.ReadString('\n')
					intentos = strings.TrimRight(intentos, "\n")
					intentosInt, _ := strconv.Atoi(intentos)
					if intentosInt <= 0 {
						fmt.Println("Error de comunicacion")
						bye = true
					} else {
						var credenciales string = ""
						for i := 0; i < intentosInt; i++ {
							fmt.Println("Por favor ingrese el nombre de usuario a iniciar sesion:")
							name, err := lector.ReadString('\n')
							name = strings.TrimRight(name, "\r\n")
							if err != nil {
							}
							fmt.Println("Por favor ingrese la contraseña para su cuenta:")
							psswd := PedirPassword()
							credenciales = name + "|" + psswd
							fmt.Println("Esperando aceptación de inicio de sesion por parte del servidor...")
							enviarLogin(&socket, &credenciales, &bye)
							if bye == true {
								break
							}
							if credenciales == "fail" {
								clearScreen()
								fmt.Println("Las credenciales de la cuenta no son correctas")
							} else {
								// segun yo aqui deberia de ir la parte de los comandos x
								clearScreen()
								fmt.Println("Inicio de sesion exitoso")

								var fin bool = false
								go enviarMensaje(&socket, &fin)
								go recibeReporte(&socket, &fin)
								for {
									if fin == true {
										bye = true
										socket.Close()
										fmt.Println("Fin de conexion")
										break
									}
								}
								break
							}
						}
						if credenciales == "fail" {
							socket.Close()
							bye = true
							break
							fmt.Println("Fin de conexion")
						}
					}

				}
				if bye {
					break
				}
			}

		}
		if bye {
			break
		}
	}
	return
}

func enviarMensaje(socketS *net.Conn, fin *bool) {
	for {
		time.Sleep(1 * time.Second)
		(*socketS).SetReadDeadline(time.Time{})
		lector := bufio.NewReader(os.Stdin)
		fmt.Println("client# digite el comando a enviar:")
		comando, _ := lector.ReadString('\n')
		env := bufio.NewWriter(*socketS)
		env.WriteString(comando + "\n")
		env.Flush()
		if comando == "bye\r\n" {

			*fin = true
			break
		}
	}
}

func enviarRegistro(socketS *net.Conn, mssg *string) {
	for {
		time.Sleep(1 * time.Second)
		env := bufio.NewWriter(*socketS)
		env.WriteString(*mssg + "\n")
		env.Flush()
		(*socketS).SetReadDeadline(time.Now().Add(5 * time.Second))
		respuesta := bufio.NewReader(*socketS)
		datoIn, tiempo := respuesta.ReadString('\n')
		if tiempo != nil {
			clearScreen()
			fmt.Println("Tiempo de espera de respuesta agotado")
			*mssg = "fin"
			break
		}
		if datoIn != "Registro exitoso\n" {
			*mssg = "fail"
			break
		}
		break
	}
}

func enviarLogin(socketS *net.Conn, credenciales *string, bye *bool) {
	for {
		time.Sleep(1 * time.Second)
		env := bufio.NewWriter(*socketS)
		env.WriteString(*credenciales + "\n")
		env.Flush()
		(*socketS).SetReadDeadline(time.Now().Add(5 * time.Second))
		respuesta := bufio.NewReader(*socketS)
		datoIn, tiempo := respuesta.ReadString('\n')
		if tiempo != nil {
			clearScreen()
			fmt.Println("Tiempo de espera de respuesta agotado")
			break
			*bye = true
		}
		if datoIn != "Login Exitoso\n" {
			*credenciales = "fail"
			break
		} else {
			break
		}
	}
}

func recibeReporte(socketS *net.Conn, fin *bool) {
	for {
		if *fin == true {
			return
		}
		var cmdLength uint16
		err := binary.Read(*socketS, binary.LittleEndian, &cmdLength)
		if err != nil {
			clearScreen()
			*fin = true
			return

		}
		cmdBytes := make([]byte, cmdLength)
		_, err = (*socketS).Read(cmdBytes)
		if err != nil {
			fmt.Println("Fin de conexion")
			return
		}
		comando := string(cmdBytes)
		fmt.Println("Client# respuesta del servidor es:", comando)
	}
}

func enviarTiempo(socketS *net.Conn, tiempo string) {
	env := bufio.NewWriter(*socketS)
	env.WriteString(tiempo + "\n")
	env.Flush()
}

func enviarOpcion(socketS *net.Conn, opcion string) {
	env := bufio.NewWriter(*socketS)
	env.WriteString(string(opcion) + "\n")
	env.Flush()
}

func PedirPassword() string {
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	return string(password)
}
