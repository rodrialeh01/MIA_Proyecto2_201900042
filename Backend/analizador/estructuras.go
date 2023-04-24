package analizador

type Partition struct {
	Part_status [1]byte
	Part_type   [1]byte
	Part_fit    [1]byte
	Part_start  int32
	Part_size   int32
	Part_name   [16]byte
}

type MBR struct {
	Mbr_tamano         int32
	Mbr_fecha_creacion [19]byte
	Mbr_dsk_signature  int32
	Mbr_fit            [1]byte
	Mbr_partition_1    Partition
	Mbr_partition_2    Partition
	Mbr_partition_3    Partition
	Mbr_partition_4    Partition
}

type EBR struct {
	Part_status [1]byte
	Part_fit    [1]byte
	Part_start  int32
	Part_size   int32
	Part_next   int32
	Part_name   [16]byte
}

type ParticionMontada struct {
	Path             string
	Name             string
	Id               string
	Type             string
	Letra            string
	Sistema_archivos bool
	Numero           int
}

var ParticionesMontadasList []ParticionMontada
