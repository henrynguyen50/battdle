const PitchleGame = {
    gameState: null,
    selectedPitchType: null,

    async init() {
        // Initialize autocomplete search
        window.PitchleGuess.init();

        // Wire Rules modal button
        const rulesBtn = document.getElementById('btn-rules');
        if (rulesBtn) {
            rulesBtn.addEventListener('click', () => {
                window.PitchleUI.showRulesModal();
            });
        }

        // Wire Rules modal close buttons
        const closeRulesBtn = document.getElementById('rules-modal-close-btn');
        if (closeRulesBtn) {
            closeRulesBtn.addEventListener('click', () => {
                window.PitchleUI.hideRulesModal();
                localStorage.setItem('pitchle_rules_seen', '1');
            });
        }

        const playRulesBtn = document.getElementById('btn-close-rules');
        if (playRulesBtn) {
            playRulesBtn.addEventListener('click', () => {
                window.PitchleUI.hideRulesModal();
                localStorage.setItem('pitchle_rules_seen', '1');
            });
        }

        // Wire result modal close button
        const closeResultBtn = document.getElementById('modal-close-btn');
        if (closeResultBtn) {
            closeResultBtn.addEventListener('click', () => {
                window.PitchleUI.hideResultModal();
            });
        }

        // Wire modal click-outside close
        window.addEventListener('click', (e) => {
            const resultModal = document.getElementById('result-modal');
            if (e.target === resultModal) {
                window.PitchleUI.hideResultModal();
            }
            const rulesModal = document.getElementById('rules-modal');
            if (e.target === rulesModal) {
                window.PitchleUI.hideRulesModal();
                localStorage.setItem('pitchle_rules_seen', '1');
            }
            const revealModal = document.getElementById('reveal-modal');
            if (e.target === revealModal) {
                window.PitchleUI.hideRevealModal();
            }
        });

        // Wire reveal modal close button
        const closeRevealBtn = document.getElementById('reveal-modal-close-btn');
        if (closeRevealBtn) {
            closeRevealBtn.addEventListener('click', () => {
                window.PitchleUI.hideRevealModal();
            });
        }

        // Wire Reveal Pitcher Test button (toggles on-page tester banner)
        const revealBtn = document.getElementById('btn-test-reveal-pitcher');
        if (revealBtn) {
            revealBtn.addEventListener('click', async () => {
                try {
                    revealBtn.disabled = true;
                    revealBtn.textContent = '⏳ Loading...';
                    const answer = await window.PitchleAPI.getPuzzleAnswerForTest();
                    window.PitchleUI.toggleTestBanner(answer);
                } catch (err) {
                    alert('Failed to get answer: ' + err.message);
                } finally {
                    revealBtn.disabled = false;
                    revealBtn.textContent = '👀 Reveal Pitcher (Test)';
                }
            });
        }

        // Wire Close Test Banner button
        const closeBannerBtn = document.getElementById('btn-close-test-banner');
        if (closeBannerBtn) {
            closeBannerBtn.addEventListener('click', () => {
                window.PitchleUI.hideTestBanner();
            });
        }
        // Wire Pitch Type Chips
        const pitchChips = document.querySelectorAll('#pitch-chips .pitch-chip');
        const submitPitchBtn = document.getElementById('btn-submit-pitch-guess');
        pitchChips.forEach(chip => {
            chip.addEventListener('click', () => {
                pitchChips.forEach(c => c.classList.remove('selected'));
                chip.classList.add('selected');
                this.selectedPitchType = chip.getAttribute('data-pitch');
                if (submitPitchBtn) {
                    submitPitchBtn.disabled = false;
                }
            });
        });

        // Wire Pitch Guess Submit button
        if (submitPitchBtn) {
            submitPitchBtn.addEventListener('click', () => {
                this.handlePitchGuessSubmit();
            });
        }

        // Wire Pitcher guess submit button
        const submitBtn = document.getElementById('btn-submit-guess');
        if (submitBtn) {
            submitBtn.addEventListener('click', () => {
                this.handleGuessSubmit();
            });
        }

        // Wire Paul Skenes Test button
        const skenesBtn = document.getElementById('btn-test-skenes');
        if (skenesBtn) {
            skenesBtn.addEventListener('click', async () => {
                try {
                    skenesBtn.disabled = true;
                    skenesBtn.textContent = '⏳ Loading Skenes...';
                    const newState = await window.PitchleAPI.setPitcherForTest(12); // Paul Skenes ID = 12
                    this.gameState = newState;
                    this.selectedPitchType = null;

                    // Clear search input and selection
                    const searchInput = document.getElementById('player-search');
                    if (searchInput) searchInput.value = '';
                    if (window.PitchleGuess) window.PitchleGuess.clearSelection();

                    // Reset pitch chip selection active states
                    const pitchChips = document.querySelectorAll('#pitch-chips .pitch-chip');
                    pitchChips.forEach(chip => chip.classList.remove('selected'));
                    const submitPitchBtn = document.getElementById('btn-submit-pitch-guess');
                    if (submitPitchBtn) {
                        submitPitchBtn.disabled = true;
                        submitPitchBtn.textContent = 'Submit Pitch Guess';
                    }

                    // Reset animation state
                    if (window.PitchleAnimation) {
                        window.PitchleAnimation.trajectoryPoints = [];
                        const replayBtn = document.getElementById('btn-replay');
                        if (replayBtn) replayBtn.disabled = true;
                    }

                    const answer = await window.PitchleAPI.getPuzzleAnswerForTest();
                    window.PitchleUI.showTestBanner(answer);
                    if (window.PitchleAnimation) {
                        window.PitchleAnimation.setPitchParams({
                            arm_angle: answer.arm_angle,
                            pitch_hand: answer.pitch_hand
                        });
                        // Automatically load and play Skenes's delivery!
                        window.PitchleAnimation.handleWatchClick();
                    }

                    this.syncUI();
                } catch (error) {
                    alert('Failed to load Paul Skenes: ' + error.message);
                } finally {
                    skenesBtn.disabled = false;
                    skenesBtn.textContent = '⚡ Paul Skenes (Test)';
                }
            });
        }
        // Wire Test Next Pitcher button
        const testBtn = document.getElementById('btn-test-next-pitcher');
        if (testBtn) {
            testBtn.addEventListener('click', async () => {
                try {
                    testBtn.disabled = true;
                    testBtn.textContent = '⏳ Loading...';
                    const newState = await window.PitchleAPI.resetPuzzleForTest();
                    this.gameState = newState;
                    this.selectedPitchType = null;

                    // Clear search input and selection
                    const searchInput = document.getElementById('player-search');
                    if (searchInput) searchInput.value = '';
                    if (window.PitchleGuess) window.PitchleGuess.clearSelection();

                    // Reset pitch chip selection active states
                    const pitchChips = document.querySelectorAll('#pitch-chips .pitch-chip');
                    pitchChips.forEach(chip => chip.classList.remove('selected'));
                    const submitPitchBtn = document.getElementById('btn-submit-pitch-guess');
                    if (submitPitchBtn) {
                        submitPitchBtn.disabled = true;
                        submitPitchBtn.textContent = 'Submit Pitch Guess';
                    }

                    // Reset animation state if applicable
                    if (window.PitchleAnimation) {
                        window.PitchleAnimation.trajectoryPoints = [];
                        const replayBtn = document.getElementById('btn-replay');
                        if (replayBtn) replayBtn.disabled = true;
                    }

                    // If test banner is visible, update it with the new answer
                    const banner = document.getElementById('test-answer-banner');
                    if (banner && banner.style.display !== 'none' && window.getComputedStyle(banner).display !== 'none') {
                        const newAnswer = await window.PitchleAPI.getPuzzleAnswerForTest();
                        window.PitchleUI.populateTestBanner(newAnswer);
                    }
                    this.syncUI();
                } catch (error) {
                    alert('Failed to reset puzzle: ' + error.message);
                } finally {
                    testBtn.disabled = false;
                    testBtn.textContent = '🎲 New Pitcher (Test)';
                }
            });
        }

        // Auto-show rules modal on first visit
        if (!localStorage.getItem('pitchle_rules_seen')) {
            window.PitchleUI.showRulesModal();
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

    async handlePitchGuessSubmit() {
        if (!this.selectedPitchType) return;

        const submitPitchBtn = document.getElementById('btn-submit-pitch-guess');
        if (submitPitchBtn) {
            submitPitchBtn.disabled = true;
            submitPitchBtn.textContent = 'Submitting...';
        }

        try {
            const updatedState = await window.PitchleAPI.submitPitchGuess(this.selectedPitchType);
            this.gameState = updatedState;
            this.syncUI();
        } catch (error) {
            alert(error.message);
            if (submitPitchBtn) {
                submitPitchBtn.disabled = false;
                submitPitchBtn.textContent = 'Submit Pitch Guess';
            }
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

        const isCompleted = this.gameState.status === 'won' || this.gameState.status === 'lost';

        // Render step 1 pitch feedback
        window.PitchleUI.renderPitchGuess(this.gameState.pitch_guess, this.gameState.pitch_guessed, isCompleted);

        // Render step 2 previous pitcher guesses
        window.PitchleUI.renderGuesses(this.gameState.guesses);

        // Render milestone hints
        const guessCount = (this.gameState.guesses || []).length;
        window.PitchleUI.renderHints(this.gameState.hints, guessCount, isCompleted);
        // Enable/disable pitcher search inputs based on phase and game completion status
        const searchInput = document.getElementById('player-search');
        const submitBtn = document.getElementById('btn-submit-guess');
        const step2Badge = document.querySelector('#pitcher-guess-section .step-badge');
        
        if (searchInput && submitBtn) {
            if (isCompleted) {
                searchInput.disabled = true;
                searchInput.placeholder = 'Game Completed!';
                submitBtn.disabled = true;
                if (step2Badge) step2Badge.classList.add('step-completed');
                
                // Show modal if not closed manually
                window.PitchleUI.showResultModal(this.gameState.status, this.gameState.answer);
            } else if (!this.gameState.pitch_guessed) {
                // Locked until Step 1 (Pitch Guess) is submitted
                searchInput.disabled = true;
                searchInput.placeholder = '🔒 Complete Step 1 (Guess Pitch) first...';
                submitBtn.disabled = true;
                if (step2Badge) step2Badge.classList.remove('step-active');
            } else {
                // Unlocked for pitcher guessing
                searchInput.disabled = false;
                searchInput.placeholder = 'Search MLB Pitchers...';
                if (step2Badge) step2Badge.classList.add('step-active');
            }
        }

        // Enable/disable 3D visualization controls
        const watchBtn = document.getElementById('btn-watch');
        if (watchBtn) {
            watchBtn.disabled = false;
        }
    }
};

// Expose globally
window.PitchleGame = PitchleGame;

// Initialize game
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        PitchleGame.init();
    });
} else {
    PitchleGame.init();
}
