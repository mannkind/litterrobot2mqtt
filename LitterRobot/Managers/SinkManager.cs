using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Threading;
using System.Threading.Channels;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.Models.Shared;
using TwoMQTT.Core;
using TwoMQTT.Core.Managers;
using TwoMQTT.Core.Models;
using MQTTnet.Extensions.ManagedClient;

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
        public SinkManager(ILogger<SinkManager> logger, IOptions<Opts> sharedOpts, IOptions<Models.SinkManager.Opts> opts,
            IManagedMqttClient client, ChannelReader<Resource> incomingData, ChannelWriter<Command> outgoingCommand) :
            base(logger, opts, client, incomingData, outgoingCommand, sharedOpts.Value.Resources, string.Empty)
        {
        }

        /// <inheritdoc />
        protected override async Task HandleIncomingMessageAsync(string topic, string payload,
            CancellationToken cancellationToken = default)
        {
            // await base.HandleIncomingMessageAsync(topic, payload, cancellationToken);
            foreach (var input in this.Questions)
            {
                var litterRobotId = this.Questions
                    .Where(x => x.Slug == input.Slug)
                    .Select(x => x.LRID)
                    .FirstOrDefault() ?? string.Empty;

                this.Logger.LogDebug($"Found {litterRobotId} for incoming data for {input.Slug}");
                if (string.IsNullOrEmpty(litterRobotId))
                {
                    this.Logger.LogDebug($"Unable to find litterRobotId for {input.Slug}");
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

                switch (topic)
                {
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.Power)):
                        cmd.Command = (int)CommandType.Power;
                        cmd.Data.Power = payload == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.Cycle)):
                        cmd.Command = (int)CommandType.Cycle;
                        cmd.Data.Cycle = payload == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.NightLightActive)):
                        cmd.Command = (int)CommandType.NightLight;
                        cmd.Data.NightLightActive = payload == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.PanelLockActive)):
                        cmd.Command = (int)CommandType.PanelLock;
                        cmd.Data.PanelLockActive = payload == Const.ON;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.CleanCycleWaitTimeMinutes)):
                        cmd.Command = (int)CommandType.WaitTime;
                        cmd.Data.CleanCycleWaitTimeMinutes = long.TryParse(payload, out var ccwtm) ? ccwtm : 0;
                        break;
                    case string s when s == this.CommandTopic(input.Slug, nameof(Resource.SleepModeActive)):
                        cmd.Command = (int)CommandType.Sleep;
                        cmd.Data.SleepMode = payload;
                        break;
                }

                this.Logger.LogDebug($"Started publishing command for litterRobitId {litterRobotId}");
                await this.OutgoingCommand.WriteAsync(cmd, cancellationToken);
                this.Logger.LogDebug($"Finished publishing command for litterRobitId {litterRobotId}");
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
                this.Logger.LogDebug($"Unable to find slug for {input.LitterRobotId}");
                return;
            }

            this.Logger.LogDebug($"Found slug {slug} for incoming data for {input.LitterRobotId}");
            this.Logger.LogDebug($"Started publishing data for slug {slug}");
            var publish = new[]
            {
                    (this.StateTopic(slug, nameof(Resource.LitterRobotId)), input.LitterRobotId),
                    (this.StateTopic(slug, nameof(Resource.PowerStatus)), input.PowerStatus),
                    (this.StateTopic(slug, nameof(Resource.UnitStatus)), input.UnitStatus),
                    (this.StateTopic(slug, nameof(Resource.UnitStatusText)), input.UnitStatusText),
                    (this.StateTopic(slug, nameof(Resource.Power)), this.BooleanOnOff(input.Power)),
                    (this.StateTopic(slug, nameof(Resource.Cycle)), this.BooleanOnOff(input.Cycle)),
                    (this.StateTopic(slug, nameof(Resource.NightLightActive)), this.BooleanOnOff(input.NightLightActive)),
                    (this.StateTopic(slug, nameof(Resource.PanelLockActive)), this.BooleanOnOff(input.PanelLockActive)),
                    (this.StateTopic(slug, nameof(Resource.DFITriggered)), this.BooleanOnOff(input.DFITriggered)),
                    (this.StateTopic(slug, nameof(Resource.SleepModeActive)), this.BooleanOnOff(input.SleepModeActive)),
                    (this.StateTopic(slug, nameof(Resource.SleepMode)), input.SleepMode),
            };
            await this.PublishAsync(publish, cancellationToken);
            this.Logger.LogDebug($"Finished publishing data for slug {slug}");
        }


        /// <inheritdoc />
        protected override IEnumerable<string> Subscriptions()
        {
            var topics = new List<string>();
            foreach (var input in this.Questions)
            {
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.Power)));
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.Cycle)));
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.NightLightActive)));
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.PanelLockActive)));
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.CleanCycleWaitTimeMinutes)));
                topics.Add(this.CommandTopic(input.Slug, nameof(Resource.SleepModeActive)));
            }

            return topics;
        }

        /// <inheritdoc />
        protected override IEnumerable<(string slug, string sensor, string type, MQTTDiscovery discovery)> Discoveries()
        {
            var discoveries = new List<(string, string, string, MQTTDiscovery)>();
            var assembly = Assembly.GetAssembly(typeof(Program))?.GetName() ?? new AssemblyName();
            var mapping = new[]
            {
                new { Sensor = nameof(Resource.LitterRobotId), Type = Const.SENSOR, Icon = "mdi:identifier" },
                new { Sensor = nameof(Resource.PowerStatus), Type = Const.SENSOR, Icon = "mdi:power-settings" },
                new { Sensor = nameof(Resource.UnitStatus), Type = Const.SENSOR, Icon = "mdi:robot" },
                new { Sensor = nameof(Resource.UnitStatusText), Type = Const.SENSOR, Icon = "mdi:robot" },
                new { Sensor = nameof(Resource.Power), Type = Const.SWITCH, Icon = "mdi:power" },
                new { Sensor = nameof(Resource.Cycle), Type = Const.SWITCH, Icon = "mdi:rotate-left" },
                new { Sensor = nameof(Resource.NightLightActive), Type = Const.SWITCH, Icon = "mdi:lightbulb" },
                new { Sensor = nameof(Resource.PanelLockActive), Type = Const.SWITCH, Icon = "mdi:lock" },
                new { Sensor = nameof(Resource.DFITriggered), Type = Const.BINARY_SENSOR, Icon = "problem" },
                new { Sensor = nameof(Resource.SleepModeActive), Type = Const.SWITCH, Icon = "mdi:sleep" },
                new { Sensor = nameof(Resource.SleepMode), Type = Const.SENSOR, Icon = "mdi:sleep" },
            };

            foreach (var input in this.Questions)
            {
                foreach (var map in mapping)
                {
                    this.Logger.LogDebug($"Generating discovery for {input.LRID} - {map.Sensor}");
                    var discovery = this.BuildDiscovery(input.Slug, map.Sensor, assembly, false);
                    if (map.Type == Const.SWITCH)
                    {
                        discovery.CommandTopic = this.CommandTopic(input.Slug, map.Sensor);
                    }

                    if (map.Type != Const.BINARY_SENSOR && !string.IsNullOrEmpty(map.Icon))
                    {
                        discovery.Icon = map.Icon;
                    }
                    else if (map.Type == Const.BINARY_SENSOR && !string.IsNullOrEmpty(map.Icon))
                    {
                        discovery.DeviceClass = map.Icon;
                    }

                    discoveries.Add((input.Slug, map.Sensor, map.Type, discovery));
                }
            }

            return discoveries;
        }
    }
}