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

            // Categories: team, division, country, height, age, debut, throws, k_percent, bb_percent, whiff, in_zone, groundballs, flyballs, popups
            const categories = [
                'team', 'division', 'country', 'height', 'age', 'debut', 'throws',
                'k_percent', 'bb_percent', 'whiff', 'in_zone', 'groundballs', 'flyballs', 'popups'
            ];
            categories.forEach(cat => {
                const td = document.createElement('td');
                const badge = document.createElement('span');
                
                const data = g.categories ? g.categories[cat] : null;
                if (data && data.value !== undefined && data.value !== null) {
                    let displayVal = data.value;
                    if (['k_percent', 'bb_percent', 'whiff', 'in_zone', 'groundballs', 'flyballs', 'popups'].includes(cat)) {
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

    showToast(message = 'Copied to clipboard! 📋') {
        const toast = document.getElementById('toast-notification');
        if (!toast) return;
        toast.textContent = message;
        toast.classList.add('show');
        clearTimeout(this._toastTimeout);
        this._toastTimeout = setTimeout(() => {
            toast.classList.remove('show');
        }, 2600);
    },

    generateShareText(gameState, answer) {
        if (!gameState) return 'Pitchle';
        const isWon = gameState.status === 'won';
        const guessCount = (gameState.guesses || []).length;
        const scoreStr = isWon ? `${guessCount}/9` : 'X/9';
        
        const now = new Date();
        const dateStr = now.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });

        let text = `Pitchle ${dateStr} ${scoreStr}\n\n`;

        // Step 1: Pitch Guess feedback
        if (gameState.pitch_guess) {
            text += `🎯 ${gameState.pitch_guess.matched ? '🟩' : '⬛'} (${gameState.pitch_guess.guessed_type || 'Pitch'})\n`;
        }

        // Step 2: Pitcher guesses (Team, Throws, Age, K%, BB%, Whiff%)
        (gameState.guesses || []).forEach(g => {
            if (!g.categories) return;
            const team = g.categories.team?.matched ? '🟩' : (g.categories.team?.close ? '🟨' : '⬛');
            const throws = g.categories.throws?.matched ? '🟩' : '⬛';
            const age = g.categories.age?.matched ? '🟩' : (g.categories.age?.close ? '🟨' : '⬛');
            const kpct = g.categories.k_percent?.matched ? '🟩' : (g.categories.k_percent?.close ? '🟨' : '⬛');
            const bbpct = g.categories.bb_percent?.matched ? '🟩' : (g.categories.bb_percent?.close ? '🟨' : '⬛');
            const whiff = g.categories.whiff?.matched ? '🟩' : (g.categories.whiff?.close ? '🟨' : '⬛');
            
            text += `${team}${throws}${age}${kpct}${bbpct}${whiff}\n`;
        });

        text += `\nPlay Pitchle at: https://pitchle.app`;
        return text;
    },

    renderGuessDistribution(distribution, userGuessCount, isWon = false) {
        const container = document.getElementById('guess-distribution-chart');
        if (!container) return;
        container.innerHTML = '';

        const rows = ['1', '2', '3', '4', '5', '6', '7', '8', '9+'];
        const dist = distribution || {};

        // Calculate max count to scale bar widths and total for percentages
        let maxCount = 0;
        let totalCount = 0;
        rows.forEach(key => {
            const count = Number(dist[key]) || 0;
            if (count > maxCount) maxCount = count;
            totalCount += count;
        });

        rows.forEach(key => {
            const count = Number(dist[key]) || 0;
            const pct = totalCount > 0 ? Math.round((count / totalCount) * 100) : 0;
            const barWidth = maxCount > 0 ? Math.max(8, Math.round((count / maxCount) * 100)) : 8;

            const rowDiv = document.createElement('div');
            rowDiv.className = 'distribution-row';

            // Check if this row is the user's winning guess row
            let isUserRow = false;
            if (isWon && userGuessCount) {
                if (key === '9+' && userGuessCount >= 9) {
                    isUserRow = true;
                } else if (String(userGuessCount) === key) {
                    isUserRow = true;
                }
            }

            if (isUserRow) {
                rowDiv.classList.add('user-guess-row');
            }

            rowDiv.innerHTML = `
                <div class="dist-label">${key}</div>
                <div class="dist-bar-track">
                    <div class="dist-bar-fill ${isUserRow ? 'highlight' : ''}" style="width: ${barWidth}%">
                        <span class="dist-bar-count">${count > 0 ? count : (isUserRow ? '1' : '0')}</span>
                    </div>
                </div>
                <div class="dist-pct">${pct}%</div>
            `;

            container.appendChild(rowDiv);
        });
    },

    showCompletionModal(gameState, stats) {
        const modal = document.getElementById('result-modal');
        if (!modal) return;

        const isWon = gameState?.status === 'won';
        const guessCount = (gameState?.guesses || []).length;
        const answer = gameState?.answer;

        // Header / Celebration Banner
        const title = document.getElementById('result-title');
        const message = document.getElementById('result-message');
        const icon = document.getElementById('celebration-icon');
        const header = document.getElementById('completion-header');

        if (isWon) {
            if (title) title.textContent = 'YOU DID IT! 🏆';
            if (message) message.textContent = `Spectacular pitching analysis! You identified the mystery pitcher in ${guessCount} guess${guessCount === 1 ? '' : 'es'}.`;
            if (icon) icon.textContent = '🏆';
            if (header) header.className = 'completion-header celebration-won';
        } else {
            if (title) title.textContent = 'GAME OVER ⚾';
            if (message) message.textContent = "Three strikes, you're out! Better luck on tomorrow's Pitchle.";
            if (icon) icon.textContent = '⚾';
            if (header) header.className = 'completion-header celebration-lost';
        }

        // Spotlight Card
        const nameEl = document.getElementById('spotlight-name');
        const teamEl = document.getElementById('spotlight-team-badge');
        const armEl = document.getElementById('spotlight-arm');
        const heightAgeEl = document.getElementById('spotlight-height-age');
        const pitchEl = document.getElementById('spotlight-pitch');
        const ratesEl = document.getElementById('spotlight-rates');
        const arsenalChips = document.getElementById('spotlight-arsenal-chips');
        const arsenalBlock = document.getElementById('spotlight-arsenal-block');

        if (answer) {
            if (nameEl) nameEl.textContent = answer.name || answer.player_name || '--';
            if (teamEl) teamEl.textContent = `${answer.team_name || ''} ${answer.division_name ? `(${answer.division_name})` : ''}`.trim() || '--';
            
            const hand = answer.pitch_hand ? `${answer.pitch_hand}HP` : (answer.throws ? `${answer.throws}HP` : 'RHP');
            const angle = answer.arm_angle !== undefined && answer.arm_angle !== null ? `${answer.arm_angle.toFixed(1)}°` : '';
            if (armEl) armEl.textContent = `${hand} ${angle}`.trim() || '--';

            const height = answer.height || '';
            const age = answer.age ? `${answer.age} yrs` : '';
            if (heightAgeEl) heightAgeEl.textContent = [height, age].filter(Boolean).join(' • ') || '--';

            const targetPitch = answer.pitch_type || (gameState?.pitch_guess ? gameState.pitch_guess.actual_type : '--');
            const velo = answer.velocity ? `${answer.velocity.toFixed(1)} mph` : (gameState?.pitch_guess ? `${gameState.pitch_guess.velocity.toFixed(1)} mph` : '');
            if (pitchEl) pitchEl.textContent = `${targetPitch} ${velo}`.trim();

            const kpct = answer.k_percent !== undefined ? `${answer.k_percent.toFixed(1)}% K` : '';
            const whiff = answer.whiff_percent !== undefined ? `${answer.whiff_percent.toFixed(1)}% Whiff` : '';
            if (ratesEl) ratesEl.textContent = [kpct, whiff].filter(Boolean).join(' / ') || '--';

            // Arsenal chips
            if (arsenalChips && arsenalBlock) {
                const pitchMix = answer.pitch_mix || (gameState?.hints ? gameState.hints.pitch_mix : []);
                if (pitchMix && pitchMix.length > 0) {
                    arsenalBlock.style.display = 'block';
                    arsenalChips.innerHTML = pitchMix.map(p => `<span class="hint-chip">${p}</span>`).join('');
                } else {
                    arsenalBlock.style.display = 'none';
                }
            }
        }

        // Solvers Banner
        const solversCountText = document.getElementById('solvers-count-text');
        if (solversCountText) {
            const totalSolved = stats?.total_solved ?? (isWon ? 1 : 0);
            if (totalSolved > 0) {
                solversCountText.textContent = `🔥 ${totalSolved.toLocaleString()} player${totalSolved === 1 ? ' has' : 's have'} solved today's Pitchle!`;
            } else {
                solversCountText.textContent = `🔥 Be the first player to solve today's Pitchle!`;
            }
        }

        // Personal Stats Widget
        const statPlayed = document.getElementById('stat-played');
        const statWinRate = document.getElementById('stat-win-rate');
        const statCurrentStreak = document.getElementById('stat-current-streak');
        const statMaxStreak = document.getElementById('stat-max-streak');

        const userStats = stats?.user_stats;
        const played = userStats?.games_played ?? 1;
        const won = userStats?.games_won ?? (isWon ? 1 : 0);
        const winRate = played > 0 ? Math.round((won / played) * 100) : (isWon ? 100 : 0);
        const currentStreak = userStats?.current_streak ?? (isWon ? 1 : 0);
        const maxStreak = userStats?.max_streak ?? (isWon ? 1 : 0);

        if (statPlayed) statPlayed.textContent = played;
        if (statWinRate) statWinRate.textContent = `${winRate}%`;
        if (statCurrentStreak) statCurrentStreak.textContent = currentStreak;
        if (statMaxStreak) statMaxStreak.textContent = maxStreak;

        // Guess Distribution Chart
        this.renderGuessDistribution(stats?.distribution, isWon ? guessCount : null, isWon);

        modal.style.display = 'flex';
    },

    showResultModal(status, answer) {
        this.showCompletionModal({ status, answer, guesses: [] }, null);
    },

    hideResultModal() {
        const modal = document.getElementById('result-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    },

    showLeaderboardModal(dailyEntries = [], streakEntries = []) {
        const modal = document.getElementById('leaderboard-modal');
        if (!modal) return;

        const formatTimeAgo = (isoStr) => {
            try {
                const d = new Date(isoStr);
                const diffMs = Date.now() - d.getTime();
                const diffMins = Math.floor(diffMs / 60000);
                if (diffMins < 1) return 'Just now';
                if (diffMins < 60) return `${diffMins}m ago`;
                const diffHours = Math.floor(diffMins / 60);
                if (diffHours < 24) return `${diffHours}h ago`;
                return d.toLocaleDateString();
            } catch (e) {
                return 'Today';
            }
        };

        // Render Daily Leaderboard
        const dailyTbody = document.getElementById('daily-leaderboard-tbody');
        const dailyEmpty = document.getElementById('daily-leaderboard-empty');

        if (dailyTbody) {
            dailyTbody.innerHTML = '';
            if (dailyEntries && dailyEntries.length > 0) {
                if (dailyEmpty) dailyEmpty.style.display = 'none';
                dailyEntries.forEach((entry, idx) => {
                    const rank = entry.rank || (idx + 1);
                    let rankBadge = `<span class="rank-num">#${rank}</span>`;
                    if (rank === 1) rankBadge = `<span class="rank-medal gold">🥇 1</span>`;
                    else if (rank === 2) rankBadge = `<span class="rank-medal silver">🥈 2</span>`;
                    else if (rank === 3) rankBadge = `<span class="rank-medal bronze">🥉 3</span>`;

                    const timeStr = entry.time_seconds ? (entry.time_seconds < 60 ? `${entry.time_seconds}s` : `${Math.floor(entry.time_seconds / 60)}m ${entry.time_seconds % 60}s`) : '--';
                    
                    const tr = document.createElement('tr');
                    if (rank <= 3) tr.className = `top-rank rank-${rank}`;
                    tr.innerHTML = `
                        <td class="col-rank">${rankBadge}</td>
                        <td class="col-player"><strong>${entry.player_name || 'Pitchle Player'}</strong></td>
                        <td class="col-guesses"><span class="badge badge-guess-count">${entry.guess_count} / 9</span></td>
                        <td class="col-pitch"><span class="badge ${entry.pitch_matched ? 'badge-match' : 'badge-miss'}">${entry.pitch_matched ? '🎯 Matched' : '⚾ Missed'}</span></td>
                        <td class="col-time">${timeStr}</td>
                        <td class="col-date">${entry.completed_at ? formatTimeAgo(entry.completed_at) : 'Today'}</td>
                    `;
                    dailyTbody.appendChild(tr);
                });
            } else {
                if (dailyEmpty) dailyEmpty.style.display = 'block';
            }
        }

        // Render Streak Leaderboard
        const streakTbody = document.getElementById('streak-leaderboard-tbody');
        const streakEmpty = document.getElementById('streak-leaderboard-empty');

        if (streakTbody) {
            streakTbody.innerHTML = '';
            if (streakEntries && streakEntries.length > 0) {
                if (streakEmpty) streakEmpty.style.display = 'none';
                streakEntries.forEach((entry, idx) => {
                    const rank = entry.rank || (idx + 1);
                    let rankBadge = `<span class="rank-num">#${rank}</span>`;
                    if (rank === 1) rankBadge = `<span class="rank-medal gold">🥇 1</span>`;
                    else if (rank === 2) rankBadge = `<span class="rank-medal silver">🥈 2</span>`;
                    else if (rank === 3) rankBadge = `<span class="rank-medal bronze">🥉 3</span>`;

                    const winRatePct = entry.win_rate !== undefined ? `${Math.round(entry.win_rate * 100)}%` : '--';

                    const tr = document.createElement('tr');
                    if (rank <= 3) tr.className = `top-rank rank-${rank}`;
                    tr.innerHTML = `
                        <td class="col-rank">${rankBadge}</td>
                        <td class="col-player"><strong>${entry.player_name || 'Pitchle Master'}</strong></td>
                        <td class="col-current"><span class="badge badge-streak">🔥 ${entry.current_streak}</span></td>
                        <td class="col-max"><span class="stat-val highlight">⚡ ${entry.max_streak}</span></td>
                        <td class="col-winrate">${winRatePct}</td>
                    `;
                    streakTbody.appendChild(tr);
                });
            } else {
                if (streakEmpty) streakEmpty.style.display = 'block';
            }
        }

        modal.style.display = 'flex';
    },

    hideLeaderboardModal() {
        const modal = document.getElementById('leaderboard-modal');
        if (modal) modal.style.display = 'none';
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
