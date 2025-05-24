import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 10 },
        { duration: '8m', target: 5 },
        { duration: '1m', target: 0 },
    ]
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    let endpoint;
    let requestName; // to differentiate requests in k6 output

    if (Math.random() < 0.4) { // 40% chance for /joke
        endpoint = '/joke';
        requestName = 'GET /joke';
    } else { // 60% chance for /random
        endpoint = '/random';
        requestName = 'GET /random';
    }

    // Make the HTTP GET request
    const res = http.get(`${BASE_URL}${endpoint}`, {
        tags: { name: requestName }, // tag the request for better metrics in k6 output
    });

    sleep(2);
}