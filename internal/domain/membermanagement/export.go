package membermanagement

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

var ErrNoExportField = errors.New("pilih minimal satu kolom untuk diexport")

var wib = time.FixedZone("WIB", 7*60*60)

type ExportInput struct {
	MemberIDs []string
	Filter    ListFilter
	Fields    []string
}

type ExportFieldOption struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Default bool   `json:"default"`
}

type exportField struct {
	Key     string
	Label   string
	Default bool
	Width   float64
	Value   func(m member.Member) interface{}
}

var statusLabels = map[string]string{
	member.StatusActive:        "Aktif",
	member.StatusUnclaimed:     "Belum Diaktivasi",
	member.StatusPendingReview: "Menunggu Persetujuan",
	member.StatusDeceased:      "Meninggal",
	member.StatusRejected:      "Ditolak",
	member.StatusIncomplete:    "Biodata Belum Lengkap",
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func dateTime(t time.Time) string {
	return t.In(wib).Format("02-01-2006 15:04")
}

func optDateTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return dateTime(*t)
}

func realEmail(email string) string {
	if strings.HasSuffix(email, "@tibta65.local") {
		return ""
	}
	return email
}

func statusLabel(s string) string {
	if label, ok := statusLabels[s]; ok {
		return label
	}
	return s
}

var exportFields = []exportField{
	{"member_number", "Nomor Induk", true, 16, func(m member.Member) interface{} { return m.MemberNumber }},
	{"full_name", "Nama Lengkap", true, 28, func(m member.Member) interface{} { return m.FullName }},
	{"nama_suci", "Nama Suci", false, 18, func(m member.Member) interface{} { return str(m.NamaSuci) }},
	{"generation", "Generasi", true, 10, func(m member.Member) interface{} { return m.Generation }},
	{"status", "Status", true, 22, func(m member.Member) interface{} { return statusLabel(m.Status) }},
	{"korda", "Korda", true, 18, func(m member.Member) interface{} { return str(m.KordaName) }},
	{"email", "Email", true, 30, func(m member.Member) interface{} { return realEmail(m.Email) }},
	{"phone", "Telepon", true, 18, func(m member.Member) interface{} { return str(m.Phone) }},
	{"address", "Alamat", false, 36, func(m member.Member) interface{} { return str(m.Address) }},
	{"username", "Username", false, 18, func(m member.Member) interface{} { return str(m.Username) }},
	{"agama", "Agama", false, 14, func(m member.Member) interface{} { return str(m.Agama) }},
	{"nrp", "NRP", false, 18, func(m member.Member) interface{} { return str(m.NRP) }},
	{"no_ak", "No AK", false, 16, func(m member.Member) interface{} { return str(m.NoAK) }},
	{"pangkat_terakhir", "Pangkat Terakhir", false, 24, func(m member.Member) interface{} { return str(m.PangkatTerakhir) }},
	{"legacy_member_number", "Nomor Induk Lama", false, 20, func(m member.Member) interface{} { return str(m.LegacyMemberNumber) }},
	{"created_at", "Tanggal Daftar", false, 20, func(m member.Member) interface{} { return dateTime(m.CreatedAt) }},
	{"approved_at", "Tanggal Disetujui", false, 20, func(m member.Member) interface{} { return optDateTime(m.ApprovedAt) }},
}

func (s *service) ExportFields() []ExportFieldOption {
	options := make([]ExportFieldOption, 0, len(exportFields))
	for _, f := range exportFields {
		options = append(options, ExportFieldOption{f.Key, f.Label, f.Default})
	}
	return options
}

func (s *service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	wanted := make(map[string]bool, len(in.Fields))
	for _, k := range in.Fields {
		wanted[k] = true
	}

	fields := make([]exportField, 0, len(in.Fields))
	for _, f := range exportFields {
		if wanted[f.Key] {
			fields = append(fields, f)
		}
	}
	if len(fields) == 0 {
		return nil, ErrNoExportField
	}

	filter := in.Filter
	if len(in.MemberIDs) > 0 {
		filter = ListFilter{IDs: in.MemberIDs}
	}

	members, err := s.repo.FindForExport(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildWorkbook(members, fields)
}

func buildWorkbook(members []member.Member, fields []exportField) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Anggota"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E5E7EB"}, Pattern: 1},
	})
	if err != nil {
		return nil, err
	}

	for i, fd := range fields {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet, col+"1", fd.Label)
		_ = f.SetColWidth(sheet, col, col, fd.Width)
	}
	lastCol, _ := excelize.ColumnNumberToName(len(fields))
	_ = f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle)

	for r, m := range members {
		for c, fd := range fields {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, fd.Value(m))
		}
	}

	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
