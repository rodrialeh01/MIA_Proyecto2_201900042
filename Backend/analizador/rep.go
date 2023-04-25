package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
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
	report := Reports{Type: "disk", Path: rep.Path, Dot: dot_generado}
	Reportes = append(Reportes, report)
	consola_rep += "[*SUCCESS*] El Reporte DISK ha sido generado con éxito. (Para poder visualizarlo es necesario iniciar sesión)\n"
}

func (rep *Rep) ReporteTree() {

}

func (rep *Rep) ReporteFile() {

}

func (rep *Rep) ReporteSB() {

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

func RetornarConsolarep() string {
	return consola_rep
}

func RetornarDot() string {
	return dot_generado
}
