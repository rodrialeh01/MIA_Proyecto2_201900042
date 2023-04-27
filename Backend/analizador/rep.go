package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Rep struct {
	Name string
	Path string
	Id   string
	Ruta string
}

var consola_rep string
var dot_generado string

func (rep *Rep) VerificarParams(parametros map[string]string) {
	consola_rep = ""
	dot_generado = ""
	//Verificando parametros obligatorios
	if rep.Name == "" {
		consola_rep += "[-ERROR-] Falta el parametro name\n"
		return
	}
	if rep.Path == "" {
		consola_rep += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if rep.Id == "" {
		consola_rep += "[-ERROR-] Falta el parametro id\n"
		return
	}

	//Verificando parametros opcionales
	if strings.ToLower(rep.Name) != "disk" && strings.ToLower(rep.Name) != "tree" && strings.ToLower(rep.Name) != "file" && strings.ToLower(rep.Name) != "sb" {
		consola_rep += "[-ERROR-] El parametro name no es valido\n"
		return
	}

	//Validando el parametro ruta
	if strings.ToLower(rep.Name) == "file" {
		if rep.Ruta == "" {
			consola_rep += "[-ERROR-] Falta el parametro ruta\n"
			return
		}
	}

	//Cambiando el path
	if rep.Path[0] == '"' && rep.Path[len(rep.Path)-1] == '"' {
		rep.Path = rep.Path[1 : len(rep.Path)-1]
	}

	//Verificando si existe el id
	if !rep.VerificarID() {
		consola_rep += "[-ERROR-] No se ha encontrado la partición con el id: " + rep.Id + "\n"
		return
	}

	switch strings.ToLower(rep.Name) {
	case "disk":
		rep.ReporteDisk()
	case "tree":
		rep.ReporteTree()
	case "file":
		rep.ReporteFile()
	case "sb":
		rep.ReporteSB()
	}

}

func (rep *Rep) ReporteDisk() {
	montada := rep.RetornarStrictMontada(rep.Id)
	if rep.IsParticionMontadaVacia(montada) {
		consola_rep += "[-ERROR-] La partición con id: " + rep.Id + " no está montada\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rep += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := rep.ObtenerParticiones(mbr)
	var ebrs []EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = rep.ListadoEBR(particiones[i], montada.Path)
			break
		}
	}
	name_disk_path := montada.Path
	//Obteniendo el nombre del disco
	if name_disk_path[0] == '"' && name_disk_path[len(name_disk_path)-1] == '"' {
		name_disk_path = name_disk_path[1 : len(name_disk_path)-1]
	}

	ruta_partida := strings.Split(name_disk_path, "/")

	//Obteniendo el nombre del disco
	name_disk := ruta_partida[len(ruta_partida)-1]

	//Creando el codigo dot

	reporte_dsk := "digraph G {\n"
	reporte_dsk += "\tlabel=\"" + name_disk + "\"; fontsize=25;\n"
	reporte_dsk += "\tnode [shape=plaintext];\n"
	reporte_dsk += "\tdisco[label=<<TABLE>"
	reporte_dsk += "\t\t<TR>\n\t\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#ffe74c\">MBR</TD>\n"
	var cero byte = 0
	fmt.Println("Tamaño: ", mbr.Mbr_tamano)
	fmt.Println("Tamaño: ", binary.Size(mbr))
	var tamanio_general = mbr.Mbr_tamano - int32(binary.Size(mbr))
	if mbr.Mbr_partition_1.Part_size > 0 {
		if string(mbr.Mbr_partition_1.Part_type[0]) == "p" || string(mbr.Mbr_partition_1.Part_type[0]) == "P" {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#f3a144\">Primaria<BR/>\n"
		} else if string(mbr.Mbr_partition_1.Part_type[0]) == "e" || string(mbr.Mbr_partition_1.Part_type[0]) == "E" {
			reporte_dsk += "\t\t<TD BGCOLOR=\"#18aa3b\">Extendida<BR/>\n"
		} else if mbr.Mbr_partition_1.Part_type[0] == cero {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#a2a2a2\">Libre<BR/>\n"
		}
		fmt.Println("1Tamaño: ", mbr.Mbr_partition_1.Part_size)
		fmt.Println("General: ", tamanio_general)
		porcentaje := float64(mbr.Mbr_partition_1.Part_size) / float64(tamanio_general)
		fmt.Println("1Porcentaje: ", porcentaje)
		dos_porcentaje := fmt.Sprintf("%.2f", porcentaje*100)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		if mbr.Mbr_partition_2.Part_size > 0 && (mbr.Mbr_partition_1.Part_size+mbr.Mbr_partition_1.Part_start) != mbr.Mbr_partition_2.Part_start {
			reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
			porcentaje = (float64(mbr.Mbr_partition_2.Part_start-(mbr.Mbr_partition_1.Part_size+mbr.Mbr_partition_1.Part_start)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
			fmt.Println("2Porcentaje: ", porcentaje)
			dos_porcentaje = fmt.Sprintf("%.2f", porcentaje)
			reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		}
	}

	if mbr.Mbr_partition_2.Part_size > 0 {
		if string(mbr.Mbr_partition_2.Part_type[0]) == "p" || string(mbr.Mbr_partition_2.Part_type[0]) == "P" {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#f3a144\">Primaria<BR/>\n"
		} else if string(mbr.Mbr_partition_2.Part_type[0]) == "e" || string(mbr.Mbr_partition_2.Part_type[0]) == "E" {
			reporte_dsk += "\t\t<TD BGCOLOR=\"#18aa3b\">Extendida<BR/>\n"
		} else if mbr.Mbr_partition_2.Part_type[0] == cero {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#a2a2a2\">Libre<BR/>\n"
		}
		porcentaje := float64(mbr.Mbr_partition_2.Part_size) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))
		fmt.Println("3Porcentaje: ", porcentaje)
		dos_porcentaje := fmt.Sprintf("%.2f", porcentaje*100)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		if mbr.Mbr_partition_3.Part_size > 0 && (mbr.Mbr_partition_2.Part_size+mbr.Mbr_partition_2.Part_start) != mbr.Mbr_partition_3.Part_start {
			reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
			porcentaje = (float64(mbr.Mbr_partition_3.Part_start-(mbr.Mbr_partition_3.Part_size+mbr.Mbr_partition_2.Part_start)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
			fmt.Println("4Porcentaje: ", porcentaje)
			dos_porcentaje = fmt.Sprintf("%.2f", porcentaje)
			reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		}
	}
	if mbr.Mbr_partition_3.Part_size > 0 {
		if string(mbr.Mbr_partition_3.Part_type[0]) == "p" || string(mbr.Mbr_partition_3.Part_type[0]) == "P" {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#f3a144\">Primaria<BR/>\n"
		} else if string(mbr.Mbr_partition_3.Part_type[0]) == "e" || string(mbr.Mbr_partition_3.Part_type[0]) == "E" {
			reporte_dsk += "\t\t<TD BGCOLOR=\"#18aa3b\">Extendida<BR/>\n"
		} else if mbr.Mbr_partition_3.Part_type[0] == cero {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#a2a2a2\">Libre<BR/>\n"
		}
		porcentaje := (float64(mbr.Mbr_partition_3.Part_size) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
		fmt.Println("14Porcentaje: ", porcentaje)
		dos_porcentaje := fmt.Sprintf("%.2f", porcentaje)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		if mbr.Mbr_partition_4.Part_size > 0 && (mbr.Mbr_partition_3.Part_size+mbr.Mbr_partition_3.Part_start) != mbr.Mbr_partition_4.Part_start {
			reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
			porcentaje = (float64(mbr.Mbr_partition_4.Part_start-(mbr.Mbr_partition_3.Part_size+mbr.Mbr_partition_3.Part_start)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
			fmt.Println("5Porcentaje: ", porcentaje)
			dos_porcentaje = fmt.Sprintf("%.2f", porcentaje)
			reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
		}
	}
	if mbr.Mbr_partition_4.Part_size > 0 {
		if string(mbr.Mbr_partition_4.Part_type[0]) == "p" || string(mbr.Mbr_partition_4.Part_type[0]) == "P" {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#f3a144\">Primaria<BR/>\n"
		} else if string(mbr.Mbr_partition_4.Part_type[0]) == "e" || string(mbr.Mbr_partition_4.Part_type[0]) == "E" {
			reporte_dsk += "\t\t<TD BGCOLOR=\"#18aa3b\">Extendida<BR/>\n"
		} else if mbr.Mbr_partition_4.Part_type[0] == cero {
			reporte_dsk += "\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"#a2a2a2\">Libre<BR/>\n"
		}
		porcentaje := (float64(mbr.Mbr_partition_4.Part_size) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
		fmt.Println("6Porcentaje: ", porcentaje)
		dos_porcentaje := fmt.Sprintf("%.2f", porcentaje)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + dos_porcentaje + "% Del disco</FONT></TD>\n"
	}

	if mbr.Mbr_partition_1.Part_size == 0 {
		reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">100% Del disco</FONT></TD>\n"
	} else if mbr.Mbr_partition_2.Part_size == 0 {
		reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
		porcentaje := (float64(mbr.Mbr_tamano-(mbr.Mbr_partition_1.Part_start+mbr.Mbr_partition_1.Part_size)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
		fmt.Println("7Porcentaje: ", porcentaje)
		porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
	} else if mbr.Mbr_partition_3.Part_size == 0 {
		reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
		porcentaje := (float64(mbr.Mbr_tamano-(mbr.Mbr_partition_2.Part_start+mbr.Mbr_partition_2.Part_size)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
		fmt.Println("8Porcentaje: ", porcentaje)
		porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
	} else if mbr.Mbr_partition_4.Part_size == 0 {
		reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
		porcentaje := (float64(mbr.Mbr_tamano-(mbr.Mbr_partition_3.Part_start+mbr.Mbr_partition_3.Part_size)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
		fmt.Println("9Porcentaje: ", porcentaje)
		porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
		reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
	} else {
		if (mbr.Mbr_tamano - (mbr.Mbr_partition_4.Part_start + mbr.Mbr_partition_4.Part_size)) != 0 {
			reporte_dsk += "\t\t<TD rowspan=\"2\" bgcolor=\"#a2a2a2\">Libre<BR/>\n"
			porcentaje := (float64(mbr.Mbr_tamano-(mbr.Mbr_partition_4.Part_start+mbr.Mbr_partition_4.Part_size)) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
			fmt.Println("10Porcentaje: ", porcentaje)
			porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
			reporte_dsk += "\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
		}
	}
	reporte_dsk += "\t</TR>\n"
	if len(ebrs) > 0 {
		reporte_dsk += "\t<TR>\n<TD>\n<TABLE BORDER=\"0\" ORDER=\"0\" CELLBORDER=\"1\" CELLPADDING=\"4\" CELLSPACING=\"0\">"
		reporte_dsk += "\t\t<TR>\n"
		fmt.Println("EBRS: ", len(ebrs))
		for i := 0; i < len(ebrs); i++ {
			if !rep.CadenaVacia(ebrs[i].Part_name) {
				reporte_dsk += "\t\t\t<TD bgcolor=\"#2fbad3\">EBR</TD>\n"
				porcentaje := (float64(ebrs[i].Part_size) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
				fmt.Println("11Porcentaje: ", porcentaje)
				porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
				reporte_dsk += "\t\t\t<TD bgcolor=\"#b9601e\">Lógica<BR/><FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
			} else {
				reporte_dsk += "\t\t\t<TD bgcolor=\"#a2a2a2\">Libre<BR/>\n"
				porcentaje := (float64(ebrs[i].Part_size) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
				fmt.Println("12Porcentaje: ", porcentaje)
				porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
				reporte_dsk += "\t\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
			}
		}
		size_max_ext := 0
		for i := 0; i < len(particiones); i++ {
			if string(particiones[i].Part_type[0]) == "e" || string(particiones[i].Part_type[0]) == "E" {
				size_max_ext = int(particiones[i].Part_start) + int(particiones[i].Part_size)
			}
		}
		if (size_max_ext - int(ebrs[len(ebrs)-1].Part_start+ebrs[len(ebrs)-1].Part_size)) != 0 {
			reporte_dsk += "\t\t\t<TD bgcolor=\"#a2a2a2\">Libre<BR/>\n"
			porcentaje := (float64(size_max_ext-(int(ebrs[len(ebrs)-1].Part_start)+int(ebrs[len(ebrs)-1].Part_size))) / float64(mbr.Mbr_tamano-int32(binary.Size(mbr)))) * 100
			fmt.Println("13Porcentaje: ", porcentaje)
			porcentaje_fs := fmt.Sprintf("%.2f", porcentaje)
			reporte_dsk += "\t\t\t<FONT POINT-SIZE=\"10\">" + porcentaje_fs + "% Del disco</FONT></TD>\n"
		}
		reporte_dsk += "\t\t</TR>\n</TABLE>\n"
		reporte_dsk += "</TD>\n\t</TR>\n"
	}
	reporte_dsk += "</TABLE>>];\n"
	reporte_dsk += "}"
	dot_generado = reporte_dsk
	fmt.Println(dot_generado)
	report := Reports{Type: "DISK", Path: rep.Path, Dot: dot_generado}
	Reportes = append(Reportes, report)
	consola_rep += "[*SUCCESS*] El Reporte DISK ha sido generado con éxito. (Para poder visualizarlo es necesario iniciar sesión)\n"
}

func (rep *Rep) ReporteTree() {
	montada := rep.RetornarStrictMontada(rep.Id)
	if rep.IsParticionMontadaVacia(montada) {
		consola_rep += "[-ERROR-] La partición con id: " + rep.Id + " no está montada\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rep += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el MBR\n"
		return
	}

	inicio_particion := 0

	particiones := rep.ObtenerParticiones(mbr)
	var ebrs []EBR
	for i := 0; i < len(particiones); i++ {
		if strings.Contains(strings.ToLower(string(particiones[i].Part_name[:])), strings.ToLower(montada.Name)) {
			inicio_particion = int(particiones[i].Part_start)
			break
		} else if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = rep.ListadoEBR(particiones[i], montada.Path)
			break
		}
	}
	for i := 0; i < len(ebrs); i++ {
		if strings.Contains(strings.ToLower(string(ebrs[i].Part_name[:])), strings.ToLower(montada.Name)) {
			inicio_particion = int(ebrs[i].Part_start)
			break
		}
	}

	if inicio_particion == 0 {
		consola_rep += "[-ERROR-] No se encontró la partición con nombre: " + rep.Id + "\n"
		return
	}

	//Leer el SuperBloque
	super_bloque := SuperBloque{}
	archivo.Seek(int64(inicio_particion), 0)
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	dot_tree := "digraph G {\n"
	dot_tree += "node [shape=plaintext]\n"
	dot_tree += "label=\"Reporte Tree\";\n"
	dot_tree += "rankdir=LR;\n"
	dot_final := ""
	dot_tree += rep.DotTree(int(super_bloque.S_inode_start), dot_final, montada.Path)
	dot_tree += "}"

	dot_generado = dot_tree
	fmt.Println(dot_generado)
	report := Reports{Type: "TREE", Path: rep.Path, Dot: dot_generado}
	Reportes = append(Reportes, report)
	consola_rep += "[*SUCCESS*] El Reporte TREE ha sido generado con éxito. (Para poder visualizarlo es necesario iniciar sesión)\n"

}

func (rep *Rep) DotTree(posicion int, dot string, path string) string {
	archivo, err := os.OpenFile(rep.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rep += "[-ERROR-] Error al abrir el archivo\n"
		return ""
	}
	defer archivo.Close()

	archivo.Seek(int64(posicion), 0)

	inode := Inodo{}
	binary.Read(archivo, binary.LittleEndian, &inode)
	str_p := strconv.Itoa(posicion)
	dot += "nodo" + str_p + " [label=<"
	dot += "<table fontname=\"Quicksand\" border=\"0\" cellspacing=\"0\">\n"
	dot += "<tr><td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"white\">Inodo</FONT></td>\n"
	dot += "<td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"#0f3fa5\">a</FONT></td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\">UID</td>\n"
	dot += "<td border=\"1\">" + string(inode.I_uid) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">GID</td>\n"
	dot += "<td border=\"1\" bgcolor=\"#9dbaf9\">" + string(inode.I_gid) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\">Size</td>\n"
	dot += "<td border=\"1\">" + string(inode.I_size) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">aTime</td>\n"
	dot += "<td border=\"1\" bgcolor=\"#9dbaf9\">" + string(inode.I_atime[:]) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\">cTime</td>\n"
	dot += "<td border=\"1\">" + string(inode.I_ctime[:]) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">mTIme</td>\n"
	dot += "<td border=\"1\" bgcolor=\"#9dbaf9\">" + string(inode.I_mtime[:]) + "</td>\n"
	dot += "</tr>\n"
	for i := 0; i < 15; i++ {
		if i%2 == 0 {
			str_i := strconv.Itoa(i + 1)
			dot += "<tr><td border=\"1\">Block " + str_i + "</td>\n"
			str_pos := strconv.Itoa(posicion)
			str_i2 := strconv.Itoa(i)
			dot += "<td border=\"1\" port=\"b" + str_pos + str_i2 + "\">" + string(inode.I_block[i]) + "</td>\n"
			dot += "</tr>\n"
		} else {
			str_i := strconv.Itoa(i + 1)
			str_pos := strconv.Itoa(posicion)
			str_i2 := strconv.Itoa(i)
			dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">Block " + str_i + "</td>\n"
			dot += "<td border=\"1\" bgcolor=\"#9dbaf9\" port=\"b" + str_pos + str_i2 + "\">" + string(inode.I_block[i]) + "</td>\n"
			dot += "</tr>\n"
		}
	}
	dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">Type</td>\n"
	str := fmt.Sprintf("%d", inode.I_type)
	dot += "<td border=\"1\" bgcolor=\"#9dbaf9\">" + string(str) + "</td>\n"
	dot += "</tr>\n"
	dot += "<tr><td border=\"1\">Perm</td>\n"
	dot += "<td border=\"1\">" + string(inode.I_perm) + "</td>\n"
	dot += "</tr>\n"
	dot += "</table>>];\n"

	for i := 0; i < 15; i++ {
		if inode.I_block[i] != -1 {
			if inode.I_type == 0 {
				str_pos := strconv.Itoa(posicion)
				str_i2 := strconv.Itoa(i)
				dot += "nodo" + str_pos + ":b" + str_pos + str_i2 + " -> nodo" + string(inode.I_block[i]) + ";\n"
				bloquec := Bloque_Carpeta{}
				archivo.Seek(int64(inode.I_block[i]), 0)
				binary.Read(archivo, binary.LittleEndian, &bloquec)
				dot += "nodo" + string(inode.I_block[i]) + "[label=<\n"
				dot += "<table fontname=\"Quicksand\" border=\"0\" cellspacing=\"0\">\n"
				dot += "<tr><td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"white\">Bloque Carpeta</FONT></td>\n"
				dot += "<td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"#0f3fa5\">a</FONT></td>\n"
				dot += "</tr>\n"
				dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">Name</td>\n"
				dot += "<td border=\"1\" bgcolor=\"#9dbaf9\"> Inodo </td>\n"
				dot += "</tr>\n"
				dot += "<tr><td border=\"1\"> . </td>\n"
				dot += "<td border=\"1\">" + string(bloquec.B_content[0].B_inodo) + "</td>\n"
				dot += "</tr>\n"
				dot += "<tr><td border=\"1\">..</td>\n"
				dot += "<td border=\"1\">" + string(bloquec.B_content[1].B_inodo) + "</td>\n"
				dot += "</tr>\n"
				name_block2 := ""
				if rep.CadenaVacia2(bloquec.B_content[2].B_name) {
					name_block2 = ""
				} else {
					name_block2 = string(bloquec.B_content[2].B_name[:])
				}
				dot += "<tr><td border=\"1\">" + name_block2 + "</td>\n"
				dos := strconv.Itoa(2)
				dot += "<td border=\"1\"  port=\"b" + string(inode.I_block[i]) + dos + "\">" + string(bloquec.B_content[2].B_inodo) + "</td>\n"
				dot += "</tr>\n"
				name_block3 := ""
				if rep.CadenaVacia2(bloquec.B_content[3].B_name) {
					name_block3 = ""
				} else {
					name_block3 = string(bloquec.B_content[3].B_name[:])
				}
				dot += "<tr><td border=\"1\">" + name_block3 + "</td>\n"
				tres := strconv.Itoa(3)
				dot += "<td border=\"1\" port=\"b" + string(inode.I_block[i]) + tres + "\">" + string(bloquec.B_content[3].B_inodo) + "</td>\n"
				dot += "</tr>\n"
				dot += "</table>>];\n"

				if bloquec.B_content[2].B_inodo != -1 {
					dot += "nodo" + string(inode.I_block[i]) + ":b" + string(inode.I_block[i]) + dos + " -> " + "nodo" + string(bloquec.B_content[2].B_inodo) + ";\n"
					dot2 := ""
					dot += rep.DotTree(int(bloquec.B_content[2].B_inodo), dot2, path)
				}

				if bloquec.B_content[3].B_inodo != -1 {
					dot += "nodo" + string(inode.I_block[i]) + ":b" + string(inode.I_block[i]) + tres + " -> " + "nodo" + string(bloquec.B_content[3].B_inodo) + ";\n"
					dot2 := ""
					dot += rep.DotTree(int(bloquec.B_content[3].B_inodo), dot2, path)
				}

			} else if inode.I_type == 1 {
				str_pos := strconv.Itoa(posicion)
				str_i2 := strconv.Itoa(i)
				dot += "nodo" + str_pos + ":b" + str_pos + str_i2 + " -> nodo" + string(inode.I_block[i]) + ";\n"
				bloquea := Bloque_Archivo{}
				archivo.Seek(int64(inode.I_block[i]), 0)
				binary.Read(archivo, binary.LittleEndian, &bloquea)
				dot += "nodo" + string(inode.I_block[i]) + "[label=<\n"
				dot += "<table fontname=\"Quicksand\" border=\"0\" cellspacing=\"0\">\n"
				dot += "<tr><td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"white\">Bloque Archivo</FONT></td>\n"
				dot += "<td bgcolor=\"#0f3fa5\" ><FONT COLOR=\"#0f3fa5\">a</FONT></td>\n"
				dot += "</tr>\n"
				dot += "<tr><td border=\"1\" bgcolor=\"#9dbaf9\">Contenido</td>\n"
				dot += "<td border=\"1\">" + string(bloquea.B_content[:]) + "</td>\n"
				dot += "</tr>\n"
				dot += "</table>>];\n"
			}
		}
	}
	return dot
}

func (rep *Rep) ReporteFile() {

}

func (rep *Rep) ReporteSB() {
	montada := rep.RetornarStrictMontada(rep.Id)
	if rep.IsParticionMontadaVacia(montada) {
		consola_rep += "[-ERROR-] La partición con id: " + rep.Id + " no está montada\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rep += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el MBR\n"
		return
	}

	inicio_particion := 0

	particiones := rep.ObtenerParticiones(mbr)
	var ebrs []EBR
	for i := 0; i < len(particiones); i++ {
		if strings.Contains(strings.ToLower(string(particiones[i].Part_name[:])), strings.ToLower(montada.Name)) {
			inicio_particion = int(particiones[i].Part_start)
			break
		} else if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = rep.ListadoEBR(particiones[i], montada.Path)
			break
		}
	}
	for i := 0; i < len(ebrs); i++ {
		if strings.Contains(strings.ToLower(string(ebrs[i].Part_name[:])), strings.ToLower(montada.Name)) {
			inicio_particion = int(ebrs[i].Part_start)
			break
		}
	}

	if inicio_particion == 0 {
		consola_rep += "[-ERROR-] No se encontró la partición con nombre: " + rep.Name + "\n"
		return
	}

	//Leer el SuperBloque
	sb := SuperBloque{}
	archivo.Seek(int64(inicio_particion), 0)
	err = binary.Read(archivo, binary.LittleEndian, &sb)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Generar el reporte
	reporte_sb := "digraph G {\n"
	reporte_sb += "node [shape=plaintext]\n"
	reporte_sb += "label=\"Reporte de SuperBloque\";\n"
	reporte_sb += "tablambr[label=<\n"
	reporte_sb += "<table fontname=\"Quicksand\" border=\"0\" cellspacing=\"0\">\n"
	reporte_sb += "<tr><td bgcolor=\"#0d7236\" ><FONT COLOR=\"white\">REPORTE DE SUPERBLOQUE</FONT></td>\n"
	reporte_sb += "<td bgcolor=\"#0d7236\" ><FONT COLOR=\"#0d7236\">a</FONT></td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_filesystem_type</td>\n"
	st_sb := strconv.Itoa(int(sb.S_filesystem_type))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_inodes_count</td>\n"
	st_sb = strconv.Itoa(int(sb.S_inodes_count))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_blocks_count</td>\n"
	st_sb = strconv.Itoa(int(sb.S_blocks_count))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_free_blocks_count</td>\n"
	st_sb = strconv.Itoa(int(sb.S_free_blocks_count))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_free_inodes_count</td>\n"
	st_sb = strconv.Itoa(int(sb.S_free_inodes_count))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_mtime</td>\n"
	st_sb = string(sb.S_mtime[:])
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_mnt_count</td>\n"
	st_sb = strconv.Itoa(int(sb.S_mnt_count))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_magic</td>\n"
	st_sb = strconv.Itoa(int(sb.S_magic))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_inode_size</td>\n"
	st_sb = strconv.Itoa(int(sb.S_inode_size))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_block_size</td>\n"
	st_sb = strconv.Itoa(int(sb.S_block_size))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_firts_ino</td>\n"
	st_sb = strconv.Itoa(int(sb.S_first_ino))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_first_blo</td>\n"
	st_sb = strconv.Itoa(int(sb.S_first_blo))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_bm_inode_start</td>\n"
	st_sb = strconv.Itoa(int(sb.S_bm_inode_start))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_bm_block_start</td>\n"
	st_sb = strconv.Itoa(int(sb.S_bm_block_start))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\" bgcolor=\"#94ffc0\">s_inode_start</td>\n"
	st_sb = strconv.Itoa(int(sb.S_inode_start))
	reporte_sb += "<td border=\"1\" bgcolor=\"#94ffc0\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "<tr><td border=\"1\">s_block_start</td>\n"
	st_sb = strconv.Itoa(int(sb.S_block_start))
	reporte_sb += "<td border=\"1\">" + st_sb + "</td>\n"
	reporte_sb += "</tr>\n"
	reporte_sb += "</table>>];\n"
	reporte_sb += "}"

	dot_generado = reporte_sb
	fmt.Println(dot_generado)
	report := Reports{Type: "SB", Path: rep.Path, Dot: dot_generado}
	Reportes = append(Reportes, report)
	consola_rep += "[*SUCCESS*] El Reporte SB ha sido generado con éxito. (Para poder visualizarlo es necesario iniciar sesión)\n"
}

func (rep *Rep) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(rep.Id)) {
			return true
		}
	}
	return false
}

func (rep *Rep) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (rep *Rep) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (rep *Rep) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (rep *Rep) ListadoEBR(Extendida Partition, path string) []EBR {
	ebrs := []EBR{}
	archivox, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer archivox.Close()

	temp := Extendida.Part_start
	for temp != -1 {
		archivox.Seek(int64(temp), 0)
		ebr := EBR{}
		err := binary.Read(archivox, binary.LittleEndian, &ebr)
		if err != nil {
			return ebrs
		}
		if !rep.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if rep.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func (rep *Rep) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (rep *Rep) CadenaVacia2(cadena [12]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func RetornarConsolarep() string {
	return consola_rep
}

func RetornarDot() string {
	return dot_generado
}
