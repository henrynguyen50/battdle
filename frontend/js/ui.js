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

    renderGuesses(guesses) {
        const tbody = document.getElementById('guesses-body');
        if (!tbody) return;

        tbody.innerHTML = '';

        guesses.forEach(g => {
            const tr = document.createElement('tr');

            // Player Name
            const tdPlayer = document.createElement('td');
            tdPlayer.textContent = g.player_name;
            tr.appendChild(tdPlayer);

            // Categories
            const categories = ['team', 'division', 'years_played', 'position', 'year_born'];
            categories.forEach(cat => {
                const td = document.createElement('td');
                const badge = document.createElement('span');
                
                const data = g.categories[cat];
                badge.textContent = data.value;
                
                if (data.matched) {
                    badge.className = 'badge badge-match';
                } else {
                    badge.className = 'badge badge-miss';
                }
                
                td.appendChild(badge);
                tr.appendChild(td);
            });

            // Result
            const tdResult = document.createElement('td');
            const resultBadge = document.createElement('span');
            resultBadge.textContent = g.result.toUpperCase();
            
            if (g.result === 'correct') {
                resultBadge.className = 'badge badge-correct';
            } else if (g.result === 'ball') {
                resultBadge.className = 'badge badge-ball';
            } else {
                resultBadge.className = 'badge badge-strike';
            }
            
            tdResult.appendChild(resultBadge);
            tr.appendChild(tdResult);

            tbody.appendChild(tr);
        });
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
    }
};

// Expose globally
window.PitchleUI = PitchleUI;
