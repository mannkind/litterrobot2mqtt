using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Threading;
using System.Threading.Channels;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using MQTTnet;
using TwoMQTT.Core.Managers;
using LitterRobot.Models.Shared;
using System.Text;
using TwoMQTT.Core;

namespace LitterRobot.Managers
{
    /// <summary>
    /// An class representing a managed way to interact with a sink.
    /// </summary>
    public class SinkManager : MQTTManager<SlugMapping, Resource, Command>
    {
        /// <summary>
        /// Initializes a new instance of the SinkManager class.
        /// </summary>
        /// <param name="logger"></param>
        /// <param name="sharedOpts"></param>
        /// <param name="opts"></param>
        /// <param name="incomingData"></param>
        /// <param name="outgoingCommand"></param>
        /// <returns></returns>
        public SinkManager(ILogger<SinkManager> logger, IOptions<Opts> sharedOpts,
            IOptions<Models.SinkManager.Opts> opts, ChannelReader<Resource> incomingData,
            ChannelWriter<Command> outgoingCommand) :
            base(logger, opts, incomingData, outgoingCommand, sharedOpts.Value.Resources)
        {
        }

        /// <inheritdoc />
        protected override async Task HandleSubscribeAsync(CancellationToken cancellationToken = default)
        {
            var tasks = new List<Task>();
            foreach (var input in this.Questions)
            {
                var topics = new List<TopicFilter>
                {
                    new TopicFilterBuilder().WithTopic(
                        $"{this.CommandTopic(input.Slug, nameof(Resource.Power))}"
                    ).Build(),
                    new TopicFilterBuilder().WithTopic(
                        $"{this.CommandTopic(input.Slug, nameof(Resource.Cycle))}"
                    ).Build(),
                    new TopicFilterBuilder().WithTopic(
                        $"{this.CommandTopic(input.Slug, nameof(Resource.NightLightActive))}"
                    ).Build(),
                    new TopicFilterBuilder().WithTopic(
                        $"{this.CommandTopic(input.Slug, nameof(Resource.PanelLockActive))}"
                    ).Build(),
                };

                tasks.Add(this.Client.SubscribeAsync(topics));
            }

            await Task.WhenAll(tasks);
        }

        /// <inheritdoc />
        protected override async Task HandleIncomingMessageAsync(MqttApplicationMessageReceivedEventArgs e,
            CancellationToken cancellationToken = default)
        {
            await base.HandleIncomingMessageAsync(e, cancellationToken);

            foreach (var input in this.Questions)
            {
                var litterRobotId = this.Questions
                    .Where(x => x.Slug == input.Slug)
                    .Select(x => x.LRID)
                    .FirstOrDefault() ?? string.Empty;

                if (string.IsNullOrEmpty(litterRobotId))
                {
                    continue;
                }

                var cmd = new Command
                {
                    Command = (int)CommandType.None,
                    Data = new Resource
                    {
                        LitterRobotId = litterRobotId,
                    }
                };

                switch (e.ApplicationMessage.Topic)
                {
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.Power)):
                        cmd.Command = (int)CommandType.Power;
                        cmd.Data.Power = Encoding.UTF8.GetString(e.ApplicationMessage.Payload) == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.Cycle)):
                        cmd.Command = (int)CommandType.Cycle;
                        cmd.Data.Cycle = Encoding.UTF8.GetString(e.ApplicationMessage.Payload) == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.NightLightActive)):
                        cmd.Command = (int)CommandType.NightLight;
                        cmd.Data.NightLightActive = Encoding.UTF8.GetString(e.ApplicationMessage.Payload) == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.PanelLockActive)):
                        cmd.Command = (int)CommandType.PanelLock;
                        cmd.Data.PanelLockActive = Encoding.UTF8.GetString(e.ApplicationMessage.Payload) == Const.ON;
                        break;
                }

                await this.OutgoingCommand.WriteAsync(cmd, cancellationToken);
            }
        }

        /// <inheritdoc />
        protected override async Task HandleIncomingDataAsync(Resource input,
            CancellationToken cancellationToken = default)
        {
            var slug = this.Questions
                .Where(x => x.LRID == input.LitterRobotId)
                .Select(x => x.Slug)
                .FirstOrDefault() ?? string.Empty;

            if (string.IsNullOrEmpty(slug))
            {
                return;
            }

            await Task.WhenAll(
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.LitterRobotId)), input.LitterRobotId,
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.PowerStatus)), input.PowerStatus,
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.UnitStatus)), input.UnitStatus,
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.UnitStatusText)), input.UnitStatusText,
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.Power)), this.BooleanOnOff(input.Power),
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.Cycle)), this.BooleanOnOff(input.Cycle),
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.NightLightActive)), this.BooleanOnOff(input.NightLightActive),
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.PanelLockActive)), this.BooleanOnOff(input.PanelLockActive),
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.DFITriggered)), this.BooleanOnOff(input.DFITriggered),
                    cancellationToken
                ),
                this.PublishAsync(
                    this.StateTopic(slug, nameof(Resource.SleepModeActive)), this.BooleanOnOff(input.SleepModeActive),
                    cancellationToken
                )
            );

        }

        /// <inheritdoc />
        protected override async Task HandleDiscoveryAsync(CancellationToken cancellationToken = default)
        {
            if (!this.Opts.DiscoveryEnabled)
            {
                return;
            }

            var tasks = new List<Task>();
            var assembly = Assembly.GetAssembly(typeof(Program))?.GetName() ?? new AssemblyName();
            var mapping = new[]
            {
                new { Sensor = nameof(Resource.LitterRobotId), Type = Const.SENSOR },
                new { Sensor = nameof(Resource.PowerStatus), Type = Const.SENSOR },
                new { Sensor = nameof(Resource.UnitStatus), Type = Const.SENSOR },
                new { Sensor = nameof(Resource.UnitStatusText), Type = Const.SENSOR },
                new { Sensor = nameof(Resource.Power), Type = Const.BINARY_SENSOR },
                new { Sensor = nameof(Resource.Cycle), Type = Const.BINARY_SENSOR },
                new { Sensor = nameof(Resource.NightLightActive), Type = Const.BINARY_SENSOR },
                new { Sensor = nameof(Resource.PanelLockActive), Type = Const.BINARY_SENSOR },
                new { Sensor = nameof(Resource.DFITriggered), Type = Const.BINARY_SENSOR },
                new { Sensor = nameof(Resource.SleepModeActive), Type = Const.BINARY_SENSOR },
            };

            foreach (var input in this.Questions)
            {
                foreach (var map in mapping)
                {
                    var discovery = this.BuildDiscovery(input.Slug, map.Sensor, assembly, false);
                    tasks.Add(this.PublishDiscoveryAsync(input.Slug, map.Sensor, map.Type, discovery, cancellationToken));
                }
            }

            await Task.WhenAll(tasks);
        }
    }
}