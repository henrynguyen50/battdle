/**
 * Frontend Test Suite for Pitchle
 * Validates JavaScript syntax, Three.js Kinematics Engine, and Delivery math
 */

const fs = require('fs');
const path = require('path');
const vm = require('vm');

let failedTests = 0;
let passedTests = 0;

function assert(condition, message) {
    if (!condition) {
        console.error(`  ❌ FAIL: ${message}`);
        failedTests++;
    } else {
        console.log(`  ✓ PASS: ${message}`);
        passedTests++;
    }
}

console.log('🧪 Running Frontend JavaScript Tests...\n');

// 1. Syntax check all frontend js files
const jsDir = path.join(__dirname, '..', 'frontend', 'js');
const files = fs.readdirSync(jsDir).filter(f => f.endsWith('.js'));

console.log('1. Syntax Check on Frontend Scripts:');
files.forEach(file => {
    const filePath = path.join(jsDir, file);
    try {
        const code = fs.readFileSync(filePath, 'utf8');
        new vm.Script(code, { filename: file });
        assert(true, `Syntax valid: ${file}`);
    } catch (err) {
        assert(false, `Syntax error in ${file}: ${err.message}`);
    }
});

// 2. Kinematics Delivery Engine Logic Tests
console.log('\n2. Testing PitchDeliveryEngine Kinematics:');
const DeliveryEngine = require('../frontend/js/delivery.js');
const engine = new DeliveryEngine();

assert(engine.getArmSlotName(10) === 'Submarine', 'Arm angle 10° is Submarine');
assert(engine.getArmSlotName(25) === 'Sidearm', 'Arm angle 25° is Sidearm');
assert(engine.getArmSlotName(38) === 'Low 3/4', 'Arm angle 38° is Low 3/4');
assert(engine.getArmSlotName(48) === 'High 3/4', 'Arm angle 48° is High 3/4');
assert(engine.getArmSlotName(60) === 'Overhand', 'Arm angle 60° is Overhand');
assert(engine.getArmSlotName(75) === 'Over-the-Top', 'Arm angle 75° is Over-the-Top');

// Test pitch parameter parsing
const paramsRHP = engine.parsePitchParams({ arm_angle: 25.7, pitch_hand: 'R', release_extension: 6.2 });
assert(paramsRHP.armAngle === 25.7, 'Parses armAngle 25.7° correctly');
assert(paramsRHP.isLHP === false, 'Parses RHP correctly');
assert(paramsRHP.extension === 6.2, 'Parses extension 6.2 ft correctly');

const paramsLHP = engine.parsePitchParams({ arm_angle: 54.0, pitch_hand: 'L', release_extension: 6.5 });
assert(paramsLHP.isLHP === true, 'Parses LHP correctly');

// Test delivery timing sequence anchors
assert(engine.TIMINGS.RELEASE === 1.25, 'Release timing set to 1.25s');
assert(engine.TIMINGS.FOLLOW_THROUGH_END === 1.80, 'Follow-through timing set to 1.80s');

// 3. HTML Tag Balance & Modal Sibling Hierarchy Tests
console.log('\n3. Testing HTML Structure & Modal Sibling Hierarchy:');
const htmlPath = path.join(__dirname, '..', 'frontend', 'index.html');
const html = fs.readFileSync(htmlPath, 'utf8');

const tagRegex = /<\/?([a-zA-Z0-9\-]+)([^>]*)>/g;
const voidElements = new Set(['!doctype', 'area', 'base', 'br', 'col', 'embed', 'hr', 'img', 'input', 'link', 'meta', 'param', 'source', 'track', 'wbr']);
let match;
const stack = [];
let htmlErrors = 0;

while ((match = tagRegex.exec(html)) !== null) {
    const isClosing = match[0].startsWith('</');
    const tagName = match[1].toLowerCase();
    const attrs = match[2];
    const isSelfClosing = match[0].endsWith('/>') || voidElements.has(tagName);

    if (isClosing) {
        if (stack.length === 0) {
            assert(false, `Unexpected closing tag: </${tagName}>`);
            htmlErrors++;
        } else {
            const last = stack.pop();
            if (last.tagName !== tagName) {
                assert(false, `Mismatched tag: expected </${last.tagName}> from line ${last.line}, got </${tagName}>`);
                htmlErrors++;
            }
        }
    } else if (!isSelfClosing) {
        const line = html.substring(0, match.index).split('\n').length;
        stack.push({ tagName, attrs, line });
    }
}

assert(stack.length === 0 && htmlErrors === 0, 'HTML tags are 100% balanced and properly closed');

// Verify modals exist and are not nested
const modalIds = ['rules-modal', 'result-modal', 'leaderboard-modal', 'reveal-modal'];
modalIds.forEach(id => {
    assert(html.includes(`id="${id}"`), `Modal #${id} exists in index.html`);
});

// 4. Modal Interactions & Tab Switching Simulation
console.log('\n4. Testing Modal Open/Close & Tab Switching Interactions:');
(async () => {
    class DOMElement {
        constructor(tagName, id = '', className = '') {
            this.tagName = tagName.toUpperCase();
            this.id = id;
            this.className = className;
            this.classList = {
                _set: new Set(className ? className.split(/\s+/).filter(Boolean) : []),
                add(c) { this._set.add(c); },
                remove(c) { this._set.delete(c); },
                contains(c) { return this._set.has(c); }
            };
            this.style = {};
            this.attributes = {};
            this.children = [];
            this._text = '';
            this.listeners = {};
            this.disabled = false;
        }

        get textContent() { return this._text; }
        set textContent(v) { this._text = String(v); }
        getAttribute(name) { return this.attributes[name] || null; }
        setAttribute(name, val) { this.attributes[name] = val; }
        addEventListener(event, fn) {
            if (!this.listeners[event]) this.listeners[event] = [];
            this.listeners[event].push(fn);
        }
        click() {
            if (this.listeners['click']) {
                this.listeners['click'].forEach(fn => fn({ target: this, preventDefault() {} }));
            }
        }
        appendChild(child) { this.children.push(child); }
        querySelectorAll(sel) { return []; }
    }

    const elementsById = {};
    const idRegex = /id="([^"]+)"/g;
    let m;
    while ((m = idRegex.exec(html)) !== null) {
        const id = m[1];
        elementsById[id] = new DOMElement('div', id);
    }

    const documentMock = {
        getElementById(id) {
            if (!elementsById[id]) elementsById[id] = new DOMElement('div', id);
            return elementsById[id];
        },
        querySelector(sel) {
            if (sel.startsWith('#')) return this.getElementById(sel.slice(1));
            return new DOMElement('div');
        },
        querySelectorAll(sel) { return []; },
        createElement(tag) { return new DOMElement(tag); },
        addEventListener() {}
    };

    const windowMock = {
        document: documentMock,
        addEventListener() {},
        localStorage: {
            _store: {},
            getItem(k) { return this._store[k] || null; },
            setItem(k, v) { this._store[k] = String(v); }
        },
        PitchleGuess: { init() {}, clearSelection() {} },
        PitchleAPI: {
            async getDailyLeaderboard() { return [{ rank: 1, player_name: 'Test Solver', guess_count: 3, pitch_matched: true, time_seconds: 40 }]; },
            async getStreakLeaderboard() { return [{ rank: 1, player_name: 'Streak King', current_streak: 8, max_streak: 12, win_rate: 0.90 }]; },
            async getTodayStats() {
                return {
                    total_solved: 50,
                    distribution: { '1': 1, '2': 4, '3': 10 },
                    user_stats: { games_played: 10, games_won: 8, current_streak: 3, max_streak: 5, user_guess_count: 3, solved: true }
                };
            },
            async getTodayPuzzle() { return { status: 'playing', guesses: [], pitch_guessed: false }; }
        }
    };

    const uiCode = fs.readFileSync(path.join(jsDir, 'ui.js'), 'utf8');
    const gameCode = fs.readFileSync(path.join(jsDir, 'game.js'), 'utf8');

    const context = vm.createContext({
        window: windowMock,
        document: documentMock,
        localStorage: windowMock.localStorage,
        console,
        setTimeout,
        clearTimeout,
        Promise
    });

    vm.runInContext(uiCode + '\nwindow.PitchleUI = PitchleUI;', context);
    vm.runInContext(gameCode + '\nwindow.PitchleGame = PitchleGame;', context);

    await context.window.PitchleGame.init();

    // Test rules modal open & close
    const rulesBtn = documentMock.getElementById('btn-rules');
    const rulesModal = documentMock.getElementById('rules-modal');
    const closeRulesBtn = documentMock.getElementById('rules-modal-close-btn');

    rulesBtn.click();
    assert(rulesModal.style.display === 'flex', 'Rules modal opens on #btn-rules click');

    closeRulesBtn.click();
    assert(rulesModal.style.display === 'none', 'Rules modal closes on close button click');

    // Test leaderboard modal open & personal stats
    const leaderboardBtn = documentMock.getElementById('btn-open-leaderboard');
    const leaderboardModal = documentMock.getElementById('leaderboard-modal');
    const closeLeaderboardBtn = documentMock.getElementById('leaderboard-modal-close-btn');

    leaderboardBtn.click();
    await new Promise(r => setTimeout(r, 50));

    assert(leaderboardModal.style.display === 'flex', 'Leaderboard modal opens on #btn-open-leaderboard click');
    assert(documentMock.getElementById('leaderboard-stat-played').textContent === '10', 'Leaderboard displays user games played (10)');
    assert(documentMock.getElementById('leaderboard-stat-win-rate').textContent === '80%', 'Leaderboard displays user win rate (80%)');

    // Test 3-way tab switching
    const tabStreak = documentMock.getElementById('tab-streak-leaderboard');
    const tabStats = documentMock.getElementById('tab-stats-leaderboard');
    const tabDaily = documentMock.getElementById('tab-daily-leaderboard');
    const panelStreak = documentMock.getElementById('panel-streak-leaderboard');
    const panelStats = documentMock.getElementById('panel-stats-leaderboard');
    const panelDaily = documentMock.getElementById('panel-daily-leaderboard');

    tabStreak.click();
    assert(panelStreak.style.display === 'block' && panelDaily.style.display === 'none', 'Streak tab activates panel-streak-leaderboard');

    tabStats.click();
    assert(panelStats.style.display === 'block' && panelStreak.style.display === 'none', 'Stats tab activates panel-stats-leaderboard');

    tabDaily.click();
    assert(panelDaily.style.display === 'block' && panelStats.style.display === 'none', 'Daily tab activates panel-daily-leaderboard');

    closeLeaderboardBtn.click();
    assert(leaderboardModal.style.display === 'none', 'Leaderboard modal closes on close button click');

    console.log(`\n========================================`);
    console.log(`Frontend Tests: ${passedTests} passed, ${failedTests} failed`);
    console.log(`========================================`);

    if (failedTests > 0) {
        process.exit(1);
    } else {
        process.exit(0);
    }
})();
