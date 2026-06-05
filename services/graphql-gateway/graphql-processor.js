module.exports = {
    $randomEmail: () => {
        return `user${Math.floor(Math.random() * 10000)}@example.com`;
    },
    $randomNumber: () => {
        return Math.floor(Math.random() * 100000);
    }
};
