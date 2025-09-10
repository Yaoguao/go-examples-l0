import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const OrderRequestDuration = new Trend('order_request_duration');
const OrderSuccessRate = new Rate('order_success_rate');
const OrderFailures = new Counter('order_failures');

export let options = {
    stages: [
        { duration: '10s', target: 100 },
        { duration: '1m', target: 200 },
        { duration: '2m', target: 500 },
        { duration: '30s', target: 100 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(99)<1000'],
    },
};

const orderUIDs = [
    'b563feb7b2b84b6329608',
    'b563feb7b2b84b6127191',
    'b563feb7b2b84b6561978',
    'b563feb7b2b84b6102309',
    'b563feb7b2b84b6722539',
    'b563feb7b2b84b6454035',
    'b563feb7b2b84b6692043',
    'b563feb7b2b84b6951821',
    'b563feb7b2b84b6119540',
];

export default function () {
    group('Order API Load Test', function () {
        const orderUID = orderUIDs[Math.floor(Math.random() * orderUIDs.length)];
        const url = `http://host.docker.internal:8081/order/${orderUID}`;

        const start = Date.now();

        let response = http.get(url, {
            tags: { name: 'GetOrder' }
        });

        const duration = Date.now() - start;
        OrderRequestDuration.add(duration);

        const success = check(response, {
            'is status 200': (r) => r.status === 200,
            'has order data': (r) => r.json('order') !== null,
            'response time < 1s': (r) => r.timings.duration < 1000
        });

        if (success) {
            OrderSuccessRate.add(true);
        } else {
            OrderSuccessRate.add(false);
            OrderFailures.add(1);
            console.log(`Failed request: ${response.status} for UID: ${orderUID}`);
        }

        sleep(0.5);
    });
}