function createConsoleSuite(name) {
    var failed = 0;

    function fail(desc, err) {
        failed++;
        console.log("FAIL: " + desc + " -- " + err);
    }

    return {
        check: function (desc, fn) {
            try {
                if (fn() === false) {
                    throw new Error("assertion returned false");
                }
                console.log("PASS: " + desc);
            } catch (e) {
                fail(desc, e);
            }
        },
        fail: fail,
        finish: function () {
            console.log((failed === 0 ? "CONSOLE_E2E_PASS " : "CONSOLE_E2E_FAIL ") + name);
        }
    };
}

function zeros(n) {
    return new Array(n + 1).join("0");
}
