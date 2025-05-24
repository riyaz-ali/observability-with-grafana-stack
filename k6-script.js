import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 15 }, // Ramp up to 15 VUs over 30 seconds
        { duration: '1m30s', target: 30 },  // Stay at 30 VUs for 1 minute 30 seconds
        { duration: '30s', target: 0 },  // Ramp down to 0 VUs over 30 seconds
    ]
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    let endpoint;
    let requestName; // to differentiate requests in k6 output

    if (Math.random() < 0.6) { // 60% chance for /joke
        endpoint = '/joke';
        requestName = 'GET /joke';
    } else { // 40% chance for /random
        endpoint = '/random';
        requestName = 'GET /random';
    }

    // Make the HTTP GET request
    const res = http.get(`${BASE_URL}${endpoint}`, {
        tags: { name: requestName }, // tag the request for better metrics in k6 output
    });

    sleep(1);
}