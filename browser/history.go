package browser

// History manages browser-style navigation with back/forward functionality
type History struct {
	current    string
	back       []string
	forward    []string
	onNavigate func(url string, recordVisit bool)
	onStatus   func(message string)
}

// NewHistory creates a new history manager
func NewHistory(start string, onNavigate func(string, bool), onStatus func(string)) *History {
	return &History{
		current:    start,
		onNavigate: onNavigate,
		onStatus:   onStatus,
	}
}

// Visit adds a page to history and clears forward history
func (h *History) Visit(page string) {
	if h.current != "" {
		h.back = append(h.back, h.current)
	}
	h.current = page
	h.forward = nil // clear forward history
}

// Back navigates to the previous page in history
func (h *History) Back() {
	if len(h.back) == 0 {
		if h.onStatus != nil {
			h.onStatus("No pages in back history.")
		}
		return
	}
	h.forward = append([]string{h.current}, h.forward...)
	h.current = h.back[len(h.back)-1]
	h.back = h.back[:len(h.back)-1]
	if h.onNavigate != nil {
		h.onNavigate(h.current, false)
	}
}

// Forward navigates to the next page in history
func (h *History) Forward() {
	if len(h.forward) == 0 {
		if h.onStatus != nil {
			h.onStatus("No pages in forward history.")
		}
		return
	}
	h.back = append(h.back, h.current)
	h.current = h.forward[0]
	h.forward = h.forward[1:]
	if h.onNavigate != nil {
		h.onNavigate(h.current, false)
	}
}

// Current returns the current URL
func (h *History) Current() string {
	return h.current
}
