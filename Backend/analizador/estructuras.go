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

type SuperBloque struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             [19]byte
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

type Inodo struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime [19]byte
	I_ctime [19]byte
	I_mtime [19]byte
	I_block [16]int32
	I_type  byte
	I_perm  int32
}

type Bloque_Carpeta struct {
	B_content [4]Content
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type Bloque_Archivo struct {
	B_content [64]byte
}

type ParticionMontada struct {
	Path             string
	Name             string
	Id               string
	Type             string
	Letra            string
	Sistema_archivos bool
	Numero           int
	Logueado         bool
	User             string
	Password         string
}

type Reports struct {
	Type string
	Path string
	Dot  string
}

var ParticionesMontadasList []ParticionMontada
var Reportes []Reports
var Idlogueado string
