const PitchleUI = {
    updateScoreboard(balls, strikes) {
        // Update dots
        const ballDots = document.querySelectorAll('#balls-indicators .indicator-dot');
        const strikeDots = document.querySelectorAll('#strikes-indicators .indicator-dot');
        const countText = document.getElementById('count-text');

        // Reset all dots
        ballDots.forEach(dot => dot.classList.remove('ball-active'));
        strikeDots.forEach(dot => dot.classList.remove('strike-active'));

        // Set active balls (max 3, but game logic stops at win/loss)
        for (let i = 0; i < balls && i < ballDots.length; i++) {
            ballDots[i].classList.add('ball-active');
        }

        // Set active strikes (max 3)
        for (let i = 0; i < strikes && i < strikeDots.length; i++) {
            strikeDots[i].classList.add('strike-active');
        }

        // Update count text (Balls - Strikes)
        if (countText) {
            countText.textContent = `${balls} - ${strikes}`;
        }
    },

    renderPitchGuess(pitchGuess, pitchGuessed, isGameOver) {
        const selectorContainer = document.getElementById('pitch-selector-container');
        const clueCard = document.getElementById('pitch-clue-card');
        const feedbackBadge = document.getElementById('pitch-guess-feedback-badge');
        const feedbackText = document.getElementById('pitch-guess-feedback-text');
        const actualTypeSpan = document.getElementById('clue-actual-type');
        const velocitySpan = document.getElementById('clue-velocity');
        const spinRateSpan = document.getElementById('clue-spin-rate');
        const step1Badge = document.querySelector('#pitch-guess-section .step-badge');

        if (pitchGuessed && pitchGuess) {
            if (selectorContainer) selectorContainer.style.display = 'none';
            if (clueCard) clueCard.style.display = 'block';
            if (step1Badge) {
                step1Badge.classList.add('step-completed');
                step1Badge.textContent = 'Step 1 ✓';
            }

            if (feedbackBadge) {
                if (pitchGuess.matched) {
                    feedbackBadge.className = 'badge badge-match';
                    feedbackBadge.textContent = 'CORRECT PITCH TYPE';
                } else {
                    feedbackBadge.className = 'badge badge-miss';
                    feedbackBadge.textContent = 'INCORRECT PITCH TYPE';
                }
            }

            if (feedbackText) {
                if (pitchGuess.matched) {
                    feedbackText.textContent = `Nice eye! You correctly identified the ${pitchGuess.actual_type}.`;
                } else {
                    feedbackText.textContent = `You guessed ${pitchGuess.guessed_type}. The pitch was actually a ${pitchGuess.actual_type}.`;
                }
            }

            if (actualTypeSpan) actualTypeSpan.textContent = pitchGuess.actual_type;
            if (velocitySpan) velocitySpan.textContent = `${pitchGuess.velocity.toFixed(1)} mph`;
            if (spinRateSpan) spinRateSpan.textContent = `${Math.round(pitchGuess.spin_rate).toLocaleString()} rpm`;
        } else {
            if (selectorContainer) selectorContainer.style.display = 'block';
            if (clueCard) clueCard.style.display = 'none';
            if (step1Badge) {
                step1Badge.classList.remove('step-completed');
                step1Badge.textContent = 'Step 1';
            }
        }
    },

    renderGuesses(guesses) {
        const tbody = document.getElementById('guesses-body');
        if (!tbody) return;

        tbody.innerHTML = '';
        if (!guesses || !Array.isArray(guesses)) return;

        // Render in reverse chronological order (newest guesses at the top)
        [...guesses].reverse().forEach(g => {
            const tr = document.createElement('tr');

            // Player Name
            const tdPlayer = document.createElement('td');
            tdPlayer.textContent = g.player_name;
            tr.appendChild(tdPlayer);

            // Statcast Categories: team, division, age, throws, k_percent, bb_percent, whiff
            const categories = ['team', 'division', 'age', 'throws', 'k_percent', 'bb_percent', 'whiff'];
            categories.forEach(cat => {
                const td = document.createElement('td');
                const badge = document.createElement('span');
                
                const data = g.categories ? g.categories[cat] : null;
                if (data && data.value !== undefined && data.value !== null) {
                    let displayVal = data.value;
                    if (cat === 'k_percent' || cat === 'bb_percent' || cat === 'whiff') {
                        const num = typeof displayVal === 'number' ? displayVal : parseFloat(displayVal);
                        displayVal = !isNaN(num) ? `${num.toFixed(1)}%` : `${displayVal}%`;
                    } else if (cat === 'throws') {
                        displayVal = `${displayVal}HP`;
                    }

                    let arrow = '';
                    if (data.direction === 'higher') {
                        arrow = ' <span class="arrow arrow-up">▲</span>';
                    } else if (data.direction === 'lower') {
                        arrow = ' <span class="arrow arrow-down">▼</span>';
                    }

                    badge.innerHTML = `${displayVal}${arrow}`;
                    
                    if (data.matched) {
                        badge.className = 'badge badge-match';
                    } else if (data.close) {
                        badge.className = 'badge badge-close';
                    } else {
                        badge.className = 'badge badge-miss';
                    }
                } else {
                    badge.textContent = '--';
                    badge.className = 'badge badge-miss';
                }
                
                td.appendChild(badge);
                tr.appendChild(td);
            });

            tbody.appendChild(tr);
        });
    },

    renderHints(hints, guessCount = 0, isGameOver = false) {
        // Pitch Mix
        const mixBadge = document.getElementById('hint-mix-badge');
        const mixContent = document.getElementById('hint-mix-content');
        if (mixBadge && mixContent) {
            if (hints && hints.pitch_mix && hints.pitch_mix.length > 0) {
                mixBadge.className = 'hint-badge unlocked';
                mixBadge.textContent = 'Unlocked';
                const chipsHtml = hints.pitch_mix.map(p => `<span class="hint-chip">${p}</span>`).join('');
                mixContent.innerHTML = `<div class="hint-chips">${chipsHtml}</div>`;
            } else {
                mixBadge.className = 'hint-badge';
                mixBadge.textContent = '3 Guesses';
                mixContent.innerHTML = `<span class="hint-locked">🔒 Make 3 guesses to unlock pitch arsenal (${guessCount}/3)</span>`;
            }
        }

        // Pitcher Role
        const roleBadge = document.getElementById('hint-role-badge');
        const roleContent = document.getElementById('hint-role-content');
        if (roleBadge && roleContent) {
            if (hints && hints.role) {
                roleBadge.className = 'hint-badge unlocked';
                roleBadge.textContent = 'Unlocked';
                roleContent.innerHTML = `<span class="hint-text highlight">${hints.role}</span>`;
            } else {
                roleBadge.className = 'hint-badge';
                roleBadge.textContent = '5 Guesses';
                roleContent.innerHTML = `<span class="hint-locked">🔒 Make 5 guesses to unlock role (${guessCount}/5)</span>`;
            }
        }

        // Career / Past Teams
        const teamsBadge = document.getElementById('hint-teams-badge');
        const teamsContent = document.getElementById('hint-teams-content');
        if (teamsBadge && teamsContent) {
            if (hints && hints.past_teams && hints.past_teams.length > 0) {
                teamsBadge.className = 'hint-badge unlocked';
                teamsBadge.textContent = 'Unlocked';
                const chipsHtml = hints.past_teams.map(t => `<span class="hint-chip team-chip">${t}</span>`).join('');
                teamsContent.innerHTML = `<div class="hint-chips">${chipsHtml}</div>`;
            } else {
                teamsBadge.className = 'hint-badge';
                teamsBadge.textContent = '5 Guesses';
                teamsContent.innerHTML = `<span class="hint-locked">🔒 Make 5 guesses to unlock career teams (${guessCount}/5)</span>`;
            }
        }
    },

    showResultModal(status, answer) {
        const modal = document.getElementById('result-modal');
        const title = document.getElementById('result-title');
        const message = document.getElementById('result-message');
        const details = document.getElementById('result-details');

        if (!modal) return;

        if (status === 'won') {
            title.textContent = 'YOU WIN!';
            title.style.color = '#2ea44f';
            message.textContent = 'Spectacular job! You guessed the daily mystery pitcher.';
        } else if (status === 'lost') {
            title.textContent = 'GAME OVER';
            title.style.color = '#da3633';
            message.textContent = 'Three strikes, you\'re out! Better luck tomorrow.';
        } else {
            // Do not show modal for active games
            return;
        }

        if (answer) {
            details.innerHTML = `
                <p>Today's Mystery Pitcher was:</p>
                <h3>${answer.name}</h3>
            `;
        } else {
            details.innerHTML = '';
        }

        modal.style.display = 'flex';
    },

    hideResultModal() {
        const modal = document.getElementById('result-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    },
    showRulesModal() {
        const modal = document.getElementById('rules-modal');
        if (modal) {
            modal.style.display = 'flex';
        }
    },

    hideRulesModal() {
        const modal = document.getElementById('rules-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    },

    showRevealModal(answer) {
        const modal = document.getElementById('reveal-modal');
        const details = document.getElementById('reveal-details');
        if (!modal || !details) return;

        if (answer) {
            details.innerHTML = `
                <div class="reveal-header">
                    <h3 class="reveal-name">${answer.player_name}</h3>
                    <span class="reveal-team-badge">${answer.team_name} (${answer.division_name})</span>
                </div>
                <div class="reveal-grid">
                    <div class="reveal-stat">
                        <span class="stat-label">Throws / Arm Angle</span>
                        <span class="stat-value">${answer.pitch_hand ? answer.pitch_hand + 'HP' : 'RHP'} ${answer.arm_angle ? `(${answer.arm_angle.toFixed(1)}°)` : ''}</span>
                    </div>
                    <div class="reveal-stat">
                        <span class="stat-label">Age</span>
                        <span class="stat-value">${answer.age} (Born ${answer.birth_year})</span>
                    </div>
                    <div class="reveal-stat">
                        <span class="stat-label">Strikeout Rate (K%)</span>
                        <span class="stat-value highlight">${answer.k_percent !== undefined ? answer.k_percent.toFixed(1) + '%' : '--'}</span>
                    </div>
                    <div class="reveal-stat">
                        <span class="stat-label">Walk Rate (BB%)</span>
                        <span class="stat-value">${answer.bb_percent !== undefined ? answer.bb_percent.toFixed(1) + '%' : '--'}</span>
                    </div>
                    <div class="reveal-stat">
                        <span class="stat-label">Whiff Rate (Whiff%)</span>
                        <span class="stat-value highlight">${answer.whiff_percent !== undefined ? answer.whiff_percent.toFixed(1) + '%' : '--'}</span>
                    </div>
                    <div class="reveal-stat">
                        <span class="stat-label">Target Pitch</span>
                        <span class="stat-value highlight">${answer.pitch_type} (${answer.velocity ? answer.velocity.toFixed(1) : '--'} mph)</span>
                    </div>
                </div>
                ${answer.pitch_mix && answer.pitch_mix.length > 0 ? `
                    <div class="reveal-section-block">
                        <span class="stat-label">Full Statcast Pitch Arsenal:</span>
                        <div class="hint-chips" style="margin-top: 6px;">
                            ${answer.pitch_mix.map(p => `<span class="hint-chip">${p}</span>`).join('')}
                        </div>
                    </div>
                ` : ''}
                ${answer.past_teams && answer.past_teams.length > 0 ? `
                    <div class="reveal-section-block">
                        <span class="stat-label">Career Teams:</span>
                        <div class="hint-chips" style="margin-top: 6px;">
                            ${answer.past_teams.map(t => `<span class="hint-chip team-chip">${t}</span>`).join('')}
                        </div>
                    </div>
                ` : ''}
            `;
        } else {
            details.innerHTML = '<p>Could not load mystery pitcher answer.</p>';
        }

        modal.style.display = 'flex';
    },

    hideRevealModal() {
        const modal = document.getElementById('reveal-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    },
    populateTestBanner(answer) {
        if (!answer) return;

        if (window.PitchleAnimation && answer.arm_angle !== undefined) {
            window.PitchleAnimation.setPitchParams({
                arm_angle: answer.arm_angle,
                pitch_hand: answer.pitch_hand
            });
        }

        const setField = (id, value) => {
            const el = document.getElementById(id);
            if (el && value !== undefined && value !== null) {
                el.textContent = value;
            }
        };

        setField('test-banner-name', answer.player_name || answer.name || '--');
        setField('test-banner-pitch', `${answer.pitch_type || '--'} (${answer.velocity ? answer.velocity.toFixed(1) : '--'} mph)`);
        setField('test-banner-team', answer.team_name || '--');
        setField('test-banner-division', answer.division_name || '--');
        setField('test-banner-age', answer.age ? `${answer.age} (Born ${answer.birth_year || '--'})` : '--');
        setField('test-banner-throws', answer.pitch_hand ? `${answer.pitch_hand}HP` : 'RHP');
        setField('test-banner-kpct', answer.k_percent !== undefined ? `${answer.k_percent.toFixed(1)}%` : '--');
        setField('test-banner-bbpct', answer.bb_percent !== undefined ? `${answer.bb_percent.toFixed(1)}%` : '--');
        setField('test-banner-whiff', answer.whiff_percent !== undefined ? `${answer.whiff_percent.toFixed(1)}%` : '--');
        setField('test-banner-inzone', answer.in_zone_percent !== undefined ? `${answer.in_zone_percent.toFixed(1)}%` : '--');
        setField('test-banner-armangle', answer.arm_angle !== undefined ? `${answer.arm_angle.toFixed(1)}°` : '--');
    },

    showTestBanner(answer) {
        const banner = document.getElementById('test-answer-banner');
        if (!banner) return;
        if (answer) {
            this.populateTestBanner(answer);
        }
        banner.style.display = 'block';
    },

    hideTestBanner() {
        const banner = document.getElementById('test-answer-banner');
        if (banner) {
            banner.style.display = 'none';
        }
    },

    toggleTestBanner(answer) {
        const banner = document.getElementById('test-answer-banner');
        if (!banner) return false;
        if (banner.style.display === 'none' || window.getComputedStyle(banner).display === 'none') {
            this.showTestBanner(answer);
            return true;
        } else {
            this.hideTestBanner();
            return false;
        }
    }
};

// Expose globally
window.PitchleUI = PitchleUI;
