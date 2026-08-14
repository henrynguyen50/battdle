const API_BASE = '/api/v1';

async function apiRequest(url, options = {}) {
    const defaultHeaders = {
        'Content-Type': 'application/json',
    };

    const config = {
        ...options,
        headers: {
            ...defaultHeaders,
            ...options.headers,
        },
        // Ensure credentials (cookies) are sent
        credentials: 'include',
    };

    try {
        const response = await fetch(`${API_BASE}${url}`, config);
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error(`API Request failed for ${url}:`, error);
        throw error;
    }
}

const PitchleAPI = {
    async getTodayPuzzle() {
        return apiRequest('/puzzle/today');
    },

    async submitGuess(playerId) {
        return apiRequest('/puzzle/today/guess', {
            method: 'POST',
            body: JSON.stringify({ player_id: playerId }),
        });
    },

    async searchPlayers(query) {
        return apiRequest(`/players/search?q=${encodeURIComponent(query)}`);
    },

    async getAnimation() {
        return apiRequest('/puzzle/today/animation');
    }
};

// Expose globally
window.PitchleAPI = PitchleAPI;
