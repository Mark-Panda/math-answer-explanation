package config

// Status 返回配置生效状态摘要（不包含 API Key 等敏感信息），用于启动时打印或健康检查
func (m *Models) Status() string {
	s := "config loaded: "
	if m.OCR.Provider != "" && m.OCR.Model != "" {
		s += "ocr=" + m.OCR.Provider + "/" + m.OCR.Model
	} else {
		s += "ocr=stub(未配置)"
	}
	s += "; "
	exp := m.LLM.Explanation
	if exp.Provider != "" && exp.Model != "" {
		s += "llm=" + exp.Provider + "/" + exp.Model
	} else {
		s += "llm=stub(未配置)"
	}
	s += "; "
	if m.Video.FFmpegBin != "" {
		s += "video=ffmpeg(" + m.Video.FFmpegBin + ")"
	} else {
		s += "video=未配置"
	}
	return s
}
