package utils

// FunctionDeclaration merepresentasikan satu "tombol" yang boleh dipencet Gemini.
type FunctionDeclaration struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

type FunctionParameters struct {
	Type       string                      `json:"type"`
	Properties map[string]FunctionProperty `json:"properties"`
	Required   []string                    `json:"required"`
}

type FunctionProperty struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Enum        []string          `json:"enum,omitempty"`
	Items       *FunctionProperty `json:"items,omitempty"`
}

// GetJagaFunctionDeclarations mengembalikan daftar lengkap "tombol" yang boleh dipencet Gemini.
func GetJagaFunctionDeclarations() []FunctionDeclaration {
	return []FunctionDeclaration{
		{
			Name:        "cari_entitas",
			Description: "Gunakan ini SETIAP KALI user menyebut nama aplikasi, situs, nomor rekening, atau nomor WhatsApp spesifik dan ingin tahu status keamanannya. JANGAN pernah menjawab status suatu nama dari pengetahuanmu sendiri - selalu panggil fungsi ini dulu.",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]FunctionProperty{
					"nama": {
						Type:        "string",
						Description: "Nama aplikasi, situs, rekening, atau WA yang ingin dicek",
					},
				},
				Required: []string{"nama"},
			},
		},
		{
			Name:        "ambil_pertanyaan_kuis",
			Description: "Gunakan ini saat kamu perlu menilai risiko secara lebih detail lewat serangkaian pertanyaan, misalnya saat hasil cari_entitas tidak pasti (waspada) atau user ingin tahu seberapa berisiko situasinya.",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]FunctionProperty{
					"skenario": {
						Type:        "string",
						Description: "Pilih 'pencegahan' kalau user BELUM mengalami kerugian (baru dapat tawaran/baru mau install), atau 'penanganan' kalau user SUDAH terlanjur install/kasih data",
						Enum:        []string{"pencegahan", "penanganan"},
					},
				},
				Required: []string{"skenario"},
			},
		},
		{
			Name:        "submit_jawaban_kuis",
			Description: "Gunakan ini setelah kamu menanyakan SEMUA pertanyaan kuis dari ambil_pertanyaan_kuis satu-satu ke user dan sudah dapat semua jawabannya (ya/tidak).",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]FunctionProperty{
					"skenario": {
						Type:        "string",
						Description: "Skenario yang sama seperti saat ambil_pertanyaan_kuis dipanggil",
						Enum:        []string{"pencegahan", "penanganan"},
					},
					"question_ids": {
						Type:        "array",
						Description: "Daftar ID pertanyaan yang dijawab 'ya' oleh user, urut sesuai urutan tanya",
						Items:       &FunctionProperty{Type: "integer"},
					},
				},
				Required: []string{"skenario", "question_ids"},
			},
		},
		{
			Name:        "submit_laporan",
			Description: "Gunakan ini HANYA setelah kamu menyusun draf laporan, SUDAH menanyakan apakah user ingin melapor secara anonim atau bersedia dikontak, DAN user SUDAH SECARA EKSPLISIT menyetujui isinya untuk dikirim. JANGAN pernah submit tanpa dua hal itu (jawaban anonim + konfirmasi kirim).",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]FunctionProperty{
					"category": {
						Type:        "string",
						Description: "Kategori laporan",
						Enum:        []string{"pinjol_ilegal", "judi_online", "investasi_bodong", "lainnya"},
					},
					"reported_name": {
						Type:        "string",
						Description: "Nama aplikasi/pihak yang dilaporkan",
					},
					"description": {
						Type:        "string",
						Description: "Ringkasan kronologi kejadian, disusun rapi dari cerita user",
					},
					"severity": {
						Type:        "string",
						Description: "Pilih 'baru_isi_data' kalau user baru sebatas mengisi data/instal aplikasi tanpa kerugian nyata. Pilih 'dana_cair' kalau dana sudah cair/ditransfer tapi belum ada ancaman. Pilih 'sudah_diteror' kalau user sudah mengalami ancaman, teror, atau penyebaran data.",
						Enum:        []string{"baru_isi_data", "dana_cair", "sudah_diteror"},
					},
					"is_anonymous": {
						Type:        "boolean",
						Description: "WAJIB ditanyakan eksplisit ke user sebelum submit: true kalau user memilih melapor secara anonim, false kalau user bersedia dikontak lebih lanjut",
					},
				},
				Required: []string{"category", "reported_name", "description", "severity", "is_anonymous"},
			},
		},
		{
			Name:        "cek_status_laporan",
			Description: "Gunakan ini kalau user memberikan nomor laporan (report_id) dan ingin tahu statusnya.",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]FunctionProperty{
					"report_id": {
						Type:        "integer",
						Description: "Nomor laporan yang ingin dicek statusnya",
					},
				},
				Required: []string{"report_id"},
			},
		},
	}
}
