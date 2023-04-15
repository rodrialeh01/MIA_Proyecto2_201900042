package analizador

type Partition struct {
	part_status [1]byte
	part_type   [1]byte
	part_fit    [1]byte
	part_start  int32
	part_size   int32
	part_name   [16]byte
}

type MBR struct {
	mbr_tamano         int32
	mbr_fecha_creacion [19]byte
	mbr_dsk_signature  int32
	mbr_fit            [1]byte
	mbr_partition_1    Partition
	mbr_partition_2    Partition
	mbr_partition_3    Partition
	mbr_partition_4    Partition
}

type EBR struct {
	part_status [1]byte
	part_fit    [1]byte
	part_start  int32
	part_size   int32
	part_next   int32
	part_name   [16]byte
}
