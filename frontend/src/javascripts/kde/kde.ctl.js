

var Controller = require('radio.controller');
var KDE = require('./kde.view');


module.exports = Controller.extend({


  events: {

    global: {
      select: 'select',
      unselect: 'unselect'
    }

  },


   /**
   * Start the view.
   *
   * @param {Object} data
   */
  initialize: function(data) {
    this.view = new KDE(data);
  },


  /**
   * Show a term KDE.
   *
   * @param {String} label
   */
  select: function(label) {
    this.view.show(label);
  },


  /**
   * Hide the KDE.
   */
  unselect: function() {
    this.view.hide();
  }


});
