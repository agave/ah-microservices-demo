const Gateway = require('/var/lib/core/js/gateway');

module.exports = {
  invoiceClient: new Gateway('invoices:80', [ 'invoice' ]),
  userClient: new Gateway('users:80', [ 'user' ])
};
