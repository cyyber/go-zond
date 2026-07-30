function createConsoleSuite(name) {
    return {
        check: function (desc, fn) {
            try {
                if (fn() === false) {
                    throw new Error("assertion returned false");
                }
            } catch (e) {
                throw new Error(desc + " -- " + e);
            }
            console.log("PASS: " + desc);
        },
        finish: function () {
            console.log("CONSOLE_E2E_PASS " + name);
        }
    };
}

function zeros(n) {
    return new Array(n + 1).join("0");
}
