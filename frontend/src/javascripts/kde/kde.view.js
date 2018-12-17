

var $ = require('jquery');
var _ = require('lodash');
var d3 = require('d3-browserify');
var Backbone = require('backbone');
var MG = require('mg');
var config = require('../config');


module.exports = Backbone.View.extend({


  el: '#kde',

  options: {
   duration: 200
  },

  initialize: function(options) {
    this.data = options;
  },
  /**
   * Render a KDE plot.
   *
   * @param {String} label
   */
  show: function(label) {

    // Show the conatiner.
    this.$el.show(this.options.duration);
    this.$el.css("color","white")
    this.$el.empty()
    this.$el.append("<b>Open ports on "+label+":</b>")
    this.$el.append("<br>")
    if (this.data.nodes[label].Ports.length > 30) {
      this.$el.append("Potential corrupted open port data.")
      return
    }
    this.$el.append(this.data.nodes[label].Ports.join(" "))
    // $.getJSON('kde/'+label+'.json', function(kde) {
    //   // Cast ISO strings -> dates.
    //   kde = _.map(kde, function(datum) {
    //     datum.date = new Date(datum.date);
    //     return datum;
    //   });

    //   var xFormat = d3.time.format('%y');
      
    //   MG.data_graphic({

    //     target: '#kde',
    //     data: kde,
    //     x_accessor: 'date',
    //     y_accessor: 'value',
    //     area: false,
    //     y_axis: false,
    //     width: 400,
    //     height: 140,
    //     left: 10,
    //     right: 10,

    //     show_years: false,
    //     xax_count: 10,
    //     xax_format: function(d) {
    //       return xFormat(d);
    //     }

    //   });

    // });

  },


  /**
   * Hide the graph.
   */
  hide: function() {
    this.$el.hide(this.options.duration);
  }


});
