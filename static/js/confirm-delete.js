(function () {
  'use strict';

  function bindConfirmDelete(form) {
    form.addEventListener('submit', function (event) {
      var message = form.getAttribute('data-confirm-message') || 'Are you sure?';
      if (!window.confirm(message)) {
        event.preventDefault();
      }
    });
  }

  document.querySelectorAll('[data-confirm-delete]').forEach(function (form) {
    bindConfirmDelete(form);
  });
})();
