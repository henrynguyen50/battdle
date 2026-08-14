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

console.log(`\n========================================`);
console.log(`Frontend Tests: ${passedTests} passed, ${failedTests} failed`);
console.log(`========================================`);

if (failedTests > 0) {
    process.exit(1);
} else {
    process.exit(0);
}
