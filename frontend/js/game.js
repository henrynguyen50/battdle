const PitchleGame = {
    gameState: null,

    async init() {
        // Initialize autocomplete search
        window.PitchleGuess.init();

        // Wire modal close button
        const closeBtn = document.getElementById('modal-close-btn');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                window.PitchleUI.hideResultModal();
            });
        }

        // Wire modal click-outside close
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('result-modal');
            if (e.target === modal) {
                window.PitchleUI.hideResultModal();
            }
        });

        // Wire guess submit button
        const submitBtn = document.getElementById('btn-submit-guess');
        if (submitBtn) {
            submitBtn.addEventListener('click', () => {
                this.handleGuessSubmit();
            });
        }

        // Load today's game state
        await this.loadGameState();
    },

    async loadGameState() {
        try {
            const state = await window.PitchleAPI.getTodayPuzzle();
            this.gameState = state;
            this.syncUI();
        } catch (error) {
            console.error('Failed to load today\'s game state:', error);
            alert('Failed to load game state. Please check if services are running.');
        }
    },

    async handleGuessSubmit() {
        const player = window.PitchleGuess.selectedPlayer;
        if (!player) return;

        try {
            const updatedState = await window.PitchleAPI.submitGuess(player.ID || player.id);
            this.gameState = updatedState;
            this.syncUI();
            
            // Clear input search
            const searchInput = document.getElementById('player-search');
            if (searchInput) {
                searchInput.value = '';
            }
            window.PitchleGuess.clearSelection();

            // Show results modal if won or lost
            if (updatedState.status === 'won' || updatedState.status === 'lost') {
                window.PitchleUI.showResultModal(updatedState.status, updatedState.answer);
            }
        } catch (error) {
            alert(error.message);
        }
    },

    syncUI() {
        if (!this.gameState) return;

        // Update score count display
        window.PitchleUI.updateScoreboard(this.gameState.balls, this.gameState.strikes);

        // Render previous guesses
        window.PitchleUI.renderGuesses(this.gameState.guesses);

        // Check if game is completed
        const isCompleted = this.gameState.status === 'won' || this.gameState.status === 'lost';
        
        // Enable/disable guess inputs based on status
        const searchInput = document.getElementById('player-search');
        const submitBtn = document.getElementById('btn-submit-guess');
        
        if (searchInput && submitBtn) {
            if (isCompleted) {
                searchInput.disabled = true;
                searchInput.placeholder = 'Game Completed!';
                submitBtn.disabled = true;
                
                // Show modal if not closed manually
                window.PitchleUI.showResultModal(this.gameState.status, this.gameState.answer);
            } else {
                searchInput.disabled = false;
                searchInput.placeholder = 'Search MLB Pitchers...';
            }
        }

        // Enable/disable 3D visualization controls
        const watchBtn = document.getElementById('btn-watch');
        if (watchBtn) {
            watchBtn.disabled = false; // Always allow watching the pitch
        }
    }
};

// Initialize game on window load
window.addEventListener('load', () => {
    PitchleGame.init();
});
