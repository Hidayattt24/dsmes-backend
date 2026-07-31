package ai_chat

import (
	"fmt"
	"regexp"
	"strings"
)

// PatientHealthContext data retrieved from backend repositories to inject into prompt.
type PatientHealthContext struct {
	Name               string
	Age                int
	Gender             string
	DiabetesType       string
	BMI                float64
	DailyCalorieTarget int
	LatestBloodSugar   string // e.g. "115 mg/dL (GDP)"
	AverageBloodSugar  string // e.g. "120 mg/dL"
	ActiveMedications  []string
	RecentActivities   []string
	RecentMeals        []string
}

// PromptBuilder constructs standardized, medical-grade system prompts for DSMES Diabetes Assistant.
type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (b *PromptBuilder) BuildSystemPrompt(ctx *PatientHealthContext) string {
	var sb strings.Builder

	sb.WriteString("Anda adalah Asisten Kesehatan Diabetes DSMES (Diabetes Self-Management Education and Support), konsultan kesehatan digital yang ramah, profesional, empatik, dan berbasis bukti medis terkini.\n\n")
	
	sb.WriteString("BATASAN TOPIK DAN BATASAN PERAN (SANGAT PENTING):\n")
	sb.WriteString("1. Anda KHUSUS dan HANYA melayani pertanyaan seputar pengelolaan Diabetes Melitus, kadar gula darah, pola makan (gizi seimbang & kalori), olahraga/aktivitas fisik, kepatuhan obat/insulin, Indeks Massa Tubuh (BMI), edukasi kesehatan diabetes, serta fitur aplikasi DSMES.\n")
	sb.WriteString("2. Jika pengguna mengajukan pertanyaan di luar topik kesehatan diabetes (seperti pemrograman, politik, film, sepak bola, kripto, game, matematika, pekerjaan rumah, pengetahuan umum, wisata, agama, atau opini pribadi), Anda WAJIB menolak secara santun dan ramah.\n")
	sb.WriteString("3. Format penolakan untuk topik di luar diabetes harus mengikuti pola santun seperti berikut:\n")
	sb.WriteString("   \"Maaf, saya merupakan Asisten AI DSMES yang dirancang khusus untuk membantu seputar pengelolaan Diabetes Melitus, seperti gula darah, pola makan, aktivitas fisik, obat, edukasi kesehatan, dan perkembangan kesehatan Anda.\n\nUntuk pertanyaan di luar topik kesehatan diabetes, saya belum dapat memberikan jawaban. Silakan ajukan pertanyaan yang berkaitan dengan kesehatan atau pengelolaan diabetes, dan saya akan dengan senang hati membantu.\"\n\n")

	sb.WriteString("ATURAN FORMATTING TEKS (SANGAT PENTING):\n")
	sb.WriteString("1. DILARANG KERAS menggunakan sintaks Markdown seperti asteris ganda (**), asteris tunggal (*), pagar (#, ##, ###), garis bawah (__), backtick (`), tabel markdown, atau link markdown.\n")
	sb.WriteString("2. Tuliskan seluruh jawaban Anda dalam TEKS POLOS (PLAIN TEXT) yang rapi dan nyaman dibaca.\n")
	sb.WriteString("3. Gunakan pemisah antar paragraf yang jelas, spasi baris yang rapi, serta penomoran biasa (contoh: 1. Menu Makanan, 2. Aktivitas Fisik) tanpa cetak tebal (bold) atau cetak miring (italic).\n\n")

	sb.WriteString("PEDOMAN TEKS & TANGGAPAN:\n")
	sb.WriteString("1. Jika pengguna mengajukan pertanyaan (seperti \"apa itu diabetes\", \"bagaimana cara menurunkan gula darah\", dll), LANGSUNG berikan penjelasan medis yang lengkap, tepat, dan edukatif tanpa sekadar salam perkenalan berulang.\n")
	sb.WriteString("2. Jika pengguna menyapa (seperti \"hai\", \"halo\"), sapa pasien dengan menyebut nama pasien dan tanyakan topik kesehatan diabetes yang ingin didiskusikan hari ini.\n")
	sb.WriteString("3. Berikan edukasi yang jelas, aman, dan mudah dipahami.\n")
	sb.WriteString("4. Jika pasien mengalami gejala darurat (seperti hipoglikemia berat < 70 mg/dL dengan pusing hebat/pingsan, atau hiperglikemia sangat tinggi), segera sarankan pasien untuk berkonsultasi langsung dengan dokter atau layanan medis darurat setempat.\n")
	sb.WriteString("5. Berikan saran yang spesifik namun selalu ingatkan pasien bahwa keputusan medis final dan dosis obat merupakan wewenang dokter spesialis mereka.\n\n")

	sb.WriteString("PROFIL DAN DATA KESEHATAN PASIEN SAAT INI:\n")
	if ctx != nil {
		if ctx.Name != "" {
			sb.WriteString(fmt.Sprintf("- Nama Pasien: %s\n", ctx.Name))
		}
		if ctx.Age > 0 {
			sb.WriteString(fmt.Sprintf("- Usia: %d tahun\n", ctx.Age))
		}
		if ctx.Gender != "" {
			sb.WriteString(fmt.Sprintf("- Jenis Kelamin: %s\n", ctx.Gender))
		}
		if ctx.DiabetesType != "" {
			sb.WriteString(fmt.Sprintf("- Tipe Diabetes: %s\n", ctx.DiabetesType))
		}
		if ctx.BMI > 0 {
			sb.WriteString(fmt.Sprintf("- Indeks Massa Tubuh (BMI): %.1f kg/m²\n", ctx.BMI))
		}
		if ctx.DailyCalorieTarget > 0 {
			sb.WriteString(fmt.Sprintf("- Target Kalori Harian: %d kcal\n", ctx.DailyCalorieTarget))
		}
		if ctx.LatestBloodSugar != "" {
			sb.WriteString(fmt.Sprintf("- Gula Darah Terakhir: %s\n", ctx.LatestBloodSugar))
		}
		if ctx.AverageBloodSugar != "" {
			sb.WriteString(fmt.Sprintf("- Rata-rata Gula Darah: %s\n", ctx.AverageBloodSugar))
		}
		if len(ctx.ActiveMedications) > 0 {
			sb.WriteString(fmt.Sprintf("- Obat/Insulin Rutin: %s\n", strings.Join(ctx.ActiveMedications, ", ")))
		}
		if len(ctx.RecentActivities) > 0 {
			sb.WriteString(fmt.Sprintf("- Aktivitas Fisik Terakhir: %s\n", strings.Join(ctx.RecentActivities, ", ")))
		}
		if len(ctx.RecentMeals) > 0 {
			sb.WriteString(fmt.Sprintf("- Catatan Makanan Terakhir: %s\n", strings.Join(ctx.RecentMeals, ", ")))
		}
	} else {
		sb.WriteString("- Data profil kesehatan pasien belum lengkap.\n")
	}

	sb.WriteString("\nGunakan data kesehatan pasien di atas untuk memberikan tanggapan teks polos yang relevan, ramah, dan personal.")
	return sb.String()
}

// SanitizeAIResponse cleans raw LLM output by stripping all markdown symbols and normalizing line breaks.
func SanitizeAIResponse(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}

	res := text

	// 1. Strip Markdown Headers (# Header, ## Header, etc.)
	reHeader := regexp.MustCompile(`(?m)^#{1,6}\s*`)
	res = reHeader.ReplaceAllString(res, "")

	// 2. Strip bold and italic markdown delimiters (**text**, *text*, __text__, _text_)
	res = strings.ReplaceAll(res, "**", "")
	res = strings.ReplaceAll(res, "__", "")
	res = strings.ReplaceAll(res, "`", "")

	// 3. Clean up bullet lists (- item or * item -> • item)
	reBullet := regexp.MustCompile(`(?m)^\s*[\-\*]\s+`)
	res = reBullet.ReplaceAllString(res, "• ")

	// 4. Remove loose asterisks or underscores used for emphasis
	res = strings.ReplaceAll(res, "*", "")
	res = strings.ReplaceAll(res, "_", "")

	// 5. Normalize multiple consecutive blank lines (max 2 newlines for clean paragraph spacing)
	reMultiNewline := regexp.MustCompile(`\n{3,}`)
	res = reMultiNewline.ReplaceAllString(res, "\n\n")

	return strings.TrimSpace(res)
}
