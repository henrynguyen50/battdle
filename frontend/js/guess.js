const PitchleGuess = {
    selectedPlayer: null,
    debounceTimeout: null,
    selectedIndex: -1,
    suggestions: [],

    init() {
        const searchInput = document.getElementById('player-search');
        const submitBtn = document.getElementById('btn-submit-guess');
        const suggestionsContainer = document.getElementById('autocomplete-suggestions');

        if (!searchInput || !submitBtn || !suggestionsContainer) return;

        // Input typing listener with debounce
        searchInput.addEventListener('input', (e) => {
            this.clearSelection();
            const query = e.target.value.trim();

            clearTimeout(this.debounceTimeout);
            if (query.length < 2) {
                this.hideSuggestions();
                return;
            }

            this.debounceTimeout = setTimeout(() => {
                this.fetchAndShowSuggestions(query);
            }, 250);
        });

        // Keydown listeners for navigation
        searchInput.addEventListener('keydown', (e) => {
            const items = suggestionsContainer.querySelectorAll('.suggestion-item');
            if (items.length === 0) return;

            if (e.key === 'ArrowDown') {
                e.preventDefault();
                this.selectedIndex = (this.selectedIndex + 1) % items.length;
                this.updateActiveItem(items);
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                this.selectedIndex = (this.selectedIndex - 1 + items.length) % items.length;
                this.updateActiveItem(items);
            } else if (e.key === 'Enter') {
                e.preventDefault();
                if (this.selectedIndex >= 0 && this.selectedIndex < items.length) {
                    this.selectPlayer(this.suggestions[this.selectedIndex]);
                }
            } else if (e.key === 'Escape') {
                this.hideSuggestions();
            }
        });

        // Hide suggestions when clicking outside
        document.addEventListener('click', (e) => {
            if (e.target !== searchInput && e.target !== suggestionsContainer) {
                this.hideSuggestions();
            }
        });

        // Focus input brings back suggestions if any
        searchInput.addEventListener('focus', () => {
            if (this.suggestions.length > 0 && searchInput.value.trim().length >= 2) {
                suggestionsContainer.style.display = 'block';
            }
        });
    },

    async fetchAndShowSuggestions(query) {
        try {
            const results = await window.PitchleAPI.searchPlayers(query);
            this.suggestions = results;
            this.renderSuggestions(results);
        } catch (error) {
            console.error('Failed to load suggestions:', error);
        }
    },

    renderSuggestions(results) {
        const container = document.getElementById('autocomplete-suggestions');
        if (!container) return;

        container.innerHTML = '';
        this.selectedIndex = -1;

        if (results.length === 0) {
            this.hideSuggestions();
            return;
        }

        results.forEach((p, idx) => {
            const div = document.createElement('div');
            div.className = 'suggestion-item';
            div.innerHTML = `<strong>${p.name}</strong> <span style="color: #8b949e; font-size: 0.85rem; margin-left: 6px;">${p.team_name ? p.team_name : (p.pitch_hand ? p.pitch_hand + 'HP' : '')}</span>`;
            div.addEventListener('click', () => {
                this.selectPlayer(p);
            });
            container.appendChild(div);
        });

        container.style.display = 'block';
    },

    updateActiveItem(items) {
        items.forEach((item, idx) => {
            if (idx === this.selectedIndex) {
                item.classList.add('selected');
                item.scrollIntoView({ block: 'nearest' });
            } else {
                item.classList.remove('selected');
            }
        });
    },

    selectPlayer(player) {
        const searchInput = document.getElementById('player-search');
        const submitBtn = document.getElementById('btn-submit-guess');

        if (!searchInput || !submitBtn) return;

        searchInput.value = player.name;
        this.selectedPlayer = player;
        submitBtn.disabled = false;
        this.hideSuggestions();
    },

    clearSelection() {
        const submitBtn = document.getElementById('btn-submit-guess');
        if (submitBtn) {
            submitBtn.disabled = true;
        }
        this.selectedPlayer = null;
    },

    hideSuggestions() {
        const container = document.getElementById('autocomplete-suggestions');
        if (container) {
            container.style.display = 'none';
        }
    }
};

// Expose globally
window.PitchleGuess = PitchleGuess;
