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

        // Wire leaderboard modal open button
        const leaderboardBtn = document.getElementById('btn-open-leaderboard');
        if (leaderboardBtn) {
            leaderboardBtn.addEventListener('click', async () => {
                try {
                    leaderboardBtn.disabled = true;
                    const origText = leaderboardBtn.innerHTML;
                    leaderboardBtn.innerHTML = '<span class="leaderboard-icon">⏳</span> Loading...';
                    
                    const [daily, streaks, stats] = await Promise.all([
                        window.PitchleAPI.getDailyLeaderboard().catch(err => {
                            console.warn('Failed to load daily leaderboard:', err);
                            return [];
                        }),
                        window.PitchleAPI.getStreakLeaderboard().catch(err => {
                            console.warn('Failed to load streaks leaderboard:', err);
                            return [];
                        }),
                        window.PitchleAPI.getTodayStats().catch(err => {
                            console.warn('Failed to load today stats:', err);
                            return null;
                        })
                    ]);

                    window.PitchleUI.showLeaderboardModal(daily, streaks, stats);
                    leaderboardBtn.innerHTML = origText;
                } catch (err) {
                    alert('Could not load leaderboard: ' + err.message);
                } finally {
                    leaderboardBtn.disabled = false;
                }
            });
        }

        // Wire leaderboard modal close button
        const closeLeaderboardBtn = document.getElementById('leaderboard-modal-close-btn');
        if (closeLeaderboardBtn) {
            closeLeaderboardBtn.addEventListener('click', () => {
                window.PitchleUI.hideLeaderboardModal();
            });
        }

        // Wire leaderboard tab switching
        const tabDaily = document.getElementById('tab-daily-leaderboard');
        const tabStreak = document.getElementById('tab-streak-leaderboard');
        const tabStats = document.getElementById('tab-stats-leaderboard');
        const panelDaily = document.getElementById('panel-daily-leaderboard');
        const panelStreak = document.getElementById('panel-streak-leaderboard');
        const panelStats = document.getElementById('panel-stats-leaderboard');

        const selectLeaderboardTab = (activeTab, activePanel) => {
            [tabDaily, tabStreak, tabStats].forEach(t => t && t.classList.remove('active'));
            [panelDaily, panelStreak, panelStats].forEach(p => p && (p.style.display = 'none'));
            if (activeTab) activeTab.classList.add('active');
            if (activePanel) activePanel.style.display = 'block';
        };

        if (tabDaily) tabDaily.addEventListener('click', () => selectLeaderboardTab(tabDaily, panelDaily));
        if (tabStreak) tabStreak.addEventListener('click', () => selectLeaderboardTab(tabStreak, panelStreak));
        if (tabStats) tabStats.addEventListener('click', () => selectLeaderboardTab(tabStats, panelStats));

        // Wire Share Score button
        const shareBtn = document.getElementById('btn-share-score');
        if (shareBtn) {
            shareBtn.addEventListener('click', async () => {
                const text = window.PitchleUI.generateShareText(this.gameState, this.gameState?.answer);
                try {
                    if (navigator.clipboard && navigator.clipboard.writeText) {
                        await navigator.clipboard.writeText(text);
                        window.PitchleUI.showToast('Copied to clipboard! 📋');
                    } else {
                        const textarea = document.createElement('textarea');
                        textarea.value = text;
                        textarea.style.position = 'fixed';
                        textarea.style.opacity = '0';
                        document.body.appendChild(textarea);
                        textarea.select();
                        document.execCommand('copy');
                        document.body.removeChild(textarea);
                        window.PitchleUI.showToast('Copied to clipboard! 📋');
                    }
                } catch (err) {
                    console.error('Failed to copy share score:', err);
                    window.PitchleUI.showToast('Failed to copy to clipboard');
                }
            });
        }

        // Wire View 3D Delivery button
        const viewDeliveryBtn = document.getElementById('btn-view-delivery');
        if (viewDeliveryBtn) {
            viewDeliveryBtn.addEventListener('click', () => {
                window.PitchleUI.hideResultModal();
                if (window.PitchleAnimation) {
                    window.PitchleAnimation.handleWatchClick();
                }
            });
        }

        // Wire New Mystery Pitcher Practice button
        const practiceBtn = document.getElementById('btn-practice-new');
        if (practiceBtn) {
            practiceBtn.addEventListener('click', async () => {
                try {
                    practiceBtn.disabled = true;
                    practiceBtn.textContent = '⏳ Loading...';
                    const newState = await window.PitchleAPI.resetPuzzleForTest();
                    this.gameState = newState;
                    this.selectedPitchType = null;
                    window.PitchleUI.hideResultModal();

                    // Clear search input and selection
                    const searchInput = document.getElementById('player-search');
                    if (searchInput) searchInput.value = '';
                    if (window.PitchleGuess) window.PitchleGuess.clearSelection();

                    // Reset pitch chip selection
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

                    this.syncUI();
                } catch (err) {
                    alert('Failed to reset puzzle for practice: ' + err.message);
                } finally {
                    practiceBtn.disabled = false;
                    practiceBtn.textContent = '🎲 New Mystery Pitcher (Practice)';
                }
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
            const leaderboardModal = document.getElementById('leaderboard-modal');
            if (e.target === leaderboardModal) {
                window.PitchleUI.hideLeaderboardModal();
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
                await this.showCompletion(updatedState);
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
                this.showCompletion(this.gameState);
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
    },

    async showCompletion(gameState) {
        let stats = null;
        try {
            stats = await window.PitchleAPI.getTodayStats();
        } catch (err) {
            console.warn('Could not fetch today statistics:', err);
        }
        window.PitchleUI.showCompletionModal(gameState, stats);
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
