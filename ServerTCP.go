package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func clearScreenII() {
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

type Config struct {
	IPs      string
	Puerto   string
	Intentos string
	Usuarios []string
}

func main() {
	for {
		configFile := "/etc/serverOperativos/config.conf"
		config, err := readConfig(configFile)
		if err != nil {
			fmt.Println("Error al leer la configuración:", err)
			return
		}

		fmt.Println("==========================")
		fmt.Println("Servidor Operativos 2023.2")
		fmt.Println("==========================")
		fmt.Println("\nMateo Gutierrez Moreno 160004815\nSantiago Cortes Hernandez 160004808\nAngel Garces Vaca 160004812\n")
		fmt.Println("\nEsperando conexion...")

		tcpAddress, err := net.ResolveTCPAddr("tcp4", ":"+config.Puerto)
		if err != nil {
			fmt.Println("Error al resolver la dirección:", err)
			return
		}
		socketserver, err := net.ListenTCP("tcp", tcpAddress)
		if err != nil {
			fmt.Println("Error al abrir el socket:", err)
			return
		}

		socket, err := socketserver.Accept()
		if err != nil {
			fmt.Println("Error al aceptar la conexión:", err)
			socketserver.Close()
			return
		}
		clearScreenII()

		fmt.Println("Server# Se ha conectado (", socket.RemoteAddr(), ")")
		tiempoEspera := recibeTiempo(&socket)

		for {
			var fin bool = false
			socket.SetReadDeadline(time.Now().Add(60 * time.Second))
			reader := bufio.NewReader(socket)
			decision, tiempo := reader.ReadString('\n')
			if tiempo != nil {
				clearScreenII()
				fmt.Println("Tiempo de espera de respuesta agotado")
				break
			}
			decision = strings.TrimRight(decision, "\n")
			if decision != "1" && decision != "2" && decision != "3" {
				continue
			}
			if decision == "3" {
				break
			}
			if decision == "1" {
				reciberegistro(&socket)
			}
			if decision == "2" {
				writer := bufio.NewWriter(socket)
				writer.WriteString(config.Intentos + "\n")
				writer.Flush()
				intentosInt, _ := strconv.Atoi(config.Intentos)
				for i := 0; i < intentosInt; i++ {
					var loginbool bool = recibeLogin(&socket)
					if loginbool {

						go recibemensaje(&socket, &fin)
						go enviarReporte(&socket, tiempoEspera)
						for {
							if fin == true {
								break
							}
						}
						clearScreenII()
						fmt.Println("Fin de conexion")
						break
					}
				}
			}
			if fin == true {
				break
			}
		}
		socket.Close()
		socketserver.Close()
	}
}

func reciberegistro(socketS *net.Conn) {
	configFile := "/etc/serverOperativos/config.conf"
	var k int = 0
	var correcto bool = true
	for {
		(*socketS).SetReadDeadline(time.Now().Add(60 * time.Second))
		reader := bufio.NewReader(*socketS)
		datoIn, tiempo := reader.ReadString('\n')
		if tiempo != nil {
			clearScreenII()
			fmt.Println("Tiempo de espera de respuesta agotado")
			break
		}
		if strings.Contains(datoIn, "|") {
			parts := strings.SplitN(datoIn, "|", 2)
			if len(parts) == 2 {
				newUser := strings.TrimSpace(parts[0])
				newPasswd := strings.TrimSpace(parts[1])
				config, err := readConfig(configFile)
				if err != nil {
					return
				}
				for {
					if k >= len(config.Usuarios) || len(config.Usuarios) == 0 {
						break
					}
					if newUser == config.Usuarios[k] {
						correcto = false
					}
					k++
				}
				if correcto == true {
					addToValues(configFile, newUser, "usuario", &correcto)
					addToUserBD(correcto, newUser, newPasswd)
				}
			}
		} else {
			correcto = false
		}
		if correcto == true {
			env := bufio.NewWriter(*socketS)
			env.WriteString("Registro exitoso" + "\n")
			env.Flush()
			break
		} else {
			env := bufio.NewWriter(*socketS)
			env.WriteString("El registro ha fallado" + "\n")
			env.Flush()
			break
		}
	}
}

func recibeLogin(socketS *net.Conn) bool {
	var correcto bool = false
	var encontrado bool = false
	for {
		(*socketS).SetReadDeadline(time.Now().Add(60 * time.Second))
		reader := bufio.NewReader(*socketS)
		datoIn, tiempo := reader.ReadString('\n')
		if tiempo != nil {
			clearScreenII()
			fmt.Println("Tiempo de espera de respuesta agotado")
			break
		}
		if strings.Contains(datoIn, "|") {
			parts := strings.SplitN(datoIn, "|", 2)
			if len(parts) == 2 {
				user := strings.TrimSpace(parts[0])
				passwd := strings.TrimSpace(parts[1])
				archivo, _ := os.ReadFile("/etc/serverOperativos/users.bd")
				sArchivo := string(archivo)
				lineas := strings.Split(sArchivo, "\n")
				for _, linea := range lineas {
					if linea == "" {
						continue
					}
					paramCuenta := strings.Split(linea, ":")
					if paramCuenta[0] == user {
						encontrado = true
						usuariodb := paramCuenta[0]
						contrasenadb := paramCuenta[1]
						hashedPss := sha256.Sum256([]byte(passwd))
						hashedPssString := hex.EncodeToString(hashedPss[:])
						if user == usuariodb && hashedPssString == contrasenadb {
							correcto = true
							env := bufio.NewWriter(*socketS)
							env.WriteString("Login Exitoso" + "\n")
							env.Flush()
							break
						} else {
							env := bufio.NewWriter(*socketS)
							env.WriteString("Login Fallido" + "\n")
							env.Flush()
							break
						}
					}
				}
				if encontrado == false {
					env := bufio.NewWriter(*socketS)
					env.WriteString("Login Fallido" + "\n")
					env.Flush()
					break
				}
				break
			}
		}
	}
	return correcto
}

func recibeTiempo(socketS *net.Conn) int {
	reader := bufio.NewReader(*socketS)
	datoTiempo, _ := reader.ReadString('\n')
	datoTiempo = strings.TrimRight(datoTiempo, "\n")
	tiempo, _ := strconv.Atoi(datoTiempo)
	return tiempo
}

func recibemensaje(socketS *net.Conn, fin *bool) {
	var vacio bool
	var count int = 0
	for {
		(*socketS).SetReadDeadline(time.Now().Add(60 * time.Second))
		vacio = false
		reader := bufio.NewReader(*socketS)
		datoIn, tiempo := reader.ReadString('\n')
		if tiempo != nil {
			clearScreenII()
			fmt.Println("Tiempo de espera de respuesta agotado")
			*fin = true
			break
		}
		fmt.Println("server# Ejecutando el comando", datoIn)
		datoIn = strings.TrimRight(datoIn, "\r\n")
		array_datosIn := strings.Fields(datoIn)
		var shell *exec.Cmd
		var stdout []byte
		if len(array_datosIn) == 0 {
			vacio = true
		} else {
			shell = exec.Command(array_datosIn[0], array_datosIn[1:]...)
			stdout, _ = shell.Output()
		}
		if vacio == true {
			count++
			fmt.Println("Server# Comando no reconocido")
			env := bufio.NewWriter(*socketS)
			env.WriteString("  Comando Desconocido o sin salida")
			env.Flush()
		} else {
			if array_datosIn[0] == "bye" {
				env := bufio.NewWriter(*socketS)
				env.WriteString("  Fin de conexion por peticion del usuario")
				env.Flush()
				*fin = true
				break
			}
			if string(stdout) == "" && vacio != true {
				count = 0
				fmt.Println("Server# Comando no reconocido")
				env := bufio.NewWriter(*socketS)
				env.WriteString("  Comando Desconocido o sin salida")
				env.Flush()
			} else {
				count = 0
				fmt.Println("Server# Enviando comando")
				env := bufio.NewWriter(*socketS)
				env.WriteString("  " + string(stdout))
				env.Flush()
			}
		}
		if count >= 3 {
			fmt.Println("Server# Comunicacion perdida despues de tres intentos, fin de conexion")
			env := bufio.NewWriter(*socketS)
			env.WriteString("  Server# Comunicacion nula tras tres intentos")
			env.Flush()
			clearScreenII()
			*fin = true
			break
		}
	}
}

func enviarReporte(socketS *net.Conn, tiempoDeEspera int) {
	for {
		time.Sleep(time.Duration(tiempoDeEspera) * time.Second)
		shell := exec.Command("bash", "-c", "vmstat 1 2 | awk 'NR==3 {print 100 - $15\"%\"}'")
		stdout, _ := shell.Output()
		cpuUsage := strings.TrimSpace(string(stdout))

		shell2 := exec.Command("bash", "-c", "free | awk '/Mem/ {print int(($3/$2)*100)}'\n")
		stdout2, _ := shell2.Output()
		ramUsage := strings.TrimSpace(string(stdout2))

		shell3 := exec.Command("bash", "-c", "df -h --output=pcent | awk 'NR==2 {print $1}'\n")
		stdout3, _ := shell3.Output()
		diskUsage := strings.TrimSpace(string(stdout3))

		var estadisticas string = "  \n  --------ESTADISTICAS SERVIDOR-------- \n  Uso de cpu: " + cpuUsage + "\n  Uso de memora  ram: " + ramUsage + "%\n  uso de disco: " + diskUsage + "\n"
		env := bufio.NewWriter(*socketS)
		env.WriteString(estadisticas)
		env.Flush()
	}
}

func readConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentKey string

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "#"); i != -1 {
			line = line[:i]
		}
		if line == "" {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				setConfigValue(&config, key, value)
			}
		} else if strings.HasSuffix(line, "{") {
			currentKey = strings.TrimSpace(strings.TrimSuffix(line, "{"))
		} else if strings.HasSuffix(line, "}") {
			currentKey = ""
		} else if currentKey != "" {
			value := strings.TrimSpace(line)
			addValueToList(&config, currentKey, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return config, err
	}
	return config, nil
}

func addValueToList(config *Config, key, value string) {
	switch key {
	case "Usuarios":
		config.Usuarios = append(config.Usuarios, value)
	}
}

func setConfigValue(config *Config, key, value string) {
	switch key {
	case "Puerto":
		config.Puerto = value
	case "Intentos":
		config.Intentos = value
	case "ServerAddress":
		config.IPs = value
	}

}

func addToValues(filename string, additionalValues string, tipo string, correcto *bool) {
	config, err := readConfig(filename)
	if err != nil {
		return
	}
	if tipo == "usuario" {
		var j int = 0
		if len(config.Usuarios) != 0 {
			for {
				if additionalValues == config.Usuarios[j] {
					*correcto = false
					j++
				} else {
					j++
				}
				if j == len(config.Usuarios) {
					break
				}
			}
		}
		if *correcto == false {
			return
		} else {
			config.Usuarios = append(config.Usuarios, additionalValues)
		}
	}
	err = writeConfig(filename, config)
	if err != nil {
		return
	}
	return
}

func addToUserBD(unic bool, usuario string, contrasena string) {
	if unic {
		hashedPss := sha256.Sum256([]byte(contrasena))
		hashedPssStr := hex.EncodeToString(hashedPss[:])
		userpsw := usuario + ":" + hashedPssStr + "\n"
		archivo, _ := os.OpenFile("/etc/serverOperativos/users.bd", os.O_APPEND|os.O_WRONLY, 0777)
		archivo.WriteString(string(userpsw))
		archivo.Close()
	} else {
		fmt.Println("Ya existe un usuario con ese nombre de usuario")
	}
}

func writeConfig(filename string, config Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	fmt.Fprintf(writer, "Puerto=%s\n", config.Puerto)
	fmt.Fprintf(writer, "Intentos=%s\n", config.Intentos)
	fmt.Fprintf(writer, "ServerAddress=%s\n", config.IPs)

	fmt.Fprintf(writer, "Usuarios{")
	for _, value := range config.Usuarios {
		fmt.Fprintf(writer, " %s", "\n"+value)
	}
	fmt.Fprint(writer, " \n}\n")
	writer.Flush()
	return nil
}
