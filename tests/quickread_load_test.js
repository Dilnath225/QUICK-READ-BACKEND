import http from 'k6/http'; // This was the fix: added 'k6/'
import { check, sleep } from 'k6';
import { FormData } from 'https://jslib.k6.io/formdata/0.0.2/index.js';

export const options = {
    stages: [
        { duration: '20s', target: 10 }, // Ramp up to 10 users
        { duration: '30s', target: 10 }, // Stay at 10 users
        { duration: '10s', target: 0 },  // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must be under 500ms
    },
};

// Ensure this matches your Gin server address
const BASE_URL = 'http://localhost:8080/api';
const binFile = open('./test_prescription.jpg', 'b');

export default function () {
    // --- 1. LOGIN ---
    const loginPayload = JSON.stringify({
        email: 'kusalibandara@gmail.com',
        password: 'kushali@123',
    });

    const params = {
        headers: { 'Content-Type': 'application/json' },
    };

    const loginRes = http.post(`${BASE_URL}/login`, loginPayload, params);

    // Validate the response structure (using your specific 'data.token' path)
    const loginPassed = check(loginRes, {
        'status is 200': (r) => r.status === 200,
        'has data object': (r) => r.json().data !== undefined,
        'has token in data': (r) => r.json().data && r.json().data.token !== undefined,
    });

    if (!loginPassed) {
        console.error(`Login failed. Status: ${loginRes.status} Body: ${loginRes.body}`);
        return;
    }

    // Extract the token correctly
    const token = loginRes.json().data.token;

    const authHeaders = {
        headers: {
            'Authorization': `Bearer ${token}`,
        }
    };

    // --- 2. PRESCRIPTION UPLOAD ---
    const fd = new FormData();
    fd.append('image', http.file(binFile, 'test_prescription.jpg', 'image/jpeg'));

    const uploadRes = http.post(`${BASE_URL}/prescriptions/upload`, fd.body(), {
        headers: {
            ...authHeaders.headers,
            'Content-Type': 'multipart/form-data; boundary=' + fd.boundary,
        },
    });

    check(uploadRes, {
        'upload successful (200/201)': (r) => r.status === 200 || r.status === 201,
    });

    sleep(1);
}